package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"log"
	"strings"
	"time"
)

// ----------------------------------------
// ADD FUNCS                               |
// ----------------------------------------

// Adds user info to UserAuth and User table
func (db *DataBase) AddUser(u *config.User) (int64, error) {
	errorMsg := "add user: %w"

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if u.PasswordHash == "" {
		return 0, fmt.Errorf(errorMsg, custom_errs.ErrInvalidPassword)
	}

	if u.Role == 0 {
		u.Role = 1
	}

	result, err := tx.ExecContext(
		ctx,
		ADD_USER,
		u.UserName,
		u.Email,
		u.PasswordHash,
		u.ProfilePic,
		u.Bio,
		u.Role,
		u.Salt,
	)

	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	if err := CatchNoRowsErr(result); err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	u.ID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	if u.ProfilePic != nil && *u.ProfilePic != "" && *u.ProfilePic != "default_pfp.jpg" {
		err = db.AddImage(tx, *u.ProfilePic)
		if err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		db.putUserInCache(*u)
		db.cache.UpdateUserCount(1)
	}
	return u.ID, err
}

func (db *DataBase) AddOAuthUser(u *config.User) (int64, error) {
	errMsg := "add auth user: %w"

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if u.Role == 0 {
		u.Role = 1
	}

	result, err := tx.ExecContext(
		ctx,
		ADD_OAUTH_USER,
		u.UserName,
		u.Email,
		u.PasswordHash,
		u.ProfilePic,
		u.Bio,
		u.Role,
		u.OAuthSub,
		u.Salt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: User.Email") {
			return 0, fmt.Errorf("%w: %v", custom_errs.ErrEmailNotUnique, err)
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: User.UserName") {
			return 0, fmt.Errorf("%w: %v", custom_errs.ErrUsernameNotUnique, err)
		}
		return 0, fmt.Errorf("exec insert: %w", err)
	}
	if err := CatchNoRowsErr(result); err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	u.ID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("can't find last id %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf(errMsg, err)
	}

	if db.UseCache {
		db.putUserInCache(*u)
		db.cache.UpdateUserCount(1)
	}
	return u.ID, nil
}

// Updates Role Column on user. Valid values are guest = 0, user = 1, moderator = 2, admin = 3
func (db *DataBase) UpdateRole(username string, adminId int64, newRole int) error {
	errorMsg := "UpdateRole: %w"

	if newRole > ADMIN || newRole < GUEST {
		return fmt.Errorf(errorMsg, "invalid role")
	}

	var notifType string
	var notifMessage string
	switch newRole {
	case MOD:
		notifType = "user-promotion"
		notifMessage = "You've been promoted to moderator! Horay!"
	case USER:
		notifType = "mod-demotion"
		notifMessage = "You've been demoted to user"
	case ADMIN:
		notifType = "mod-promotion"
		notifMessage = "You've somehow become an administrator?!"
	default:
		return fmt.Errorf(errorMsg, custom_errs.ErrInvalidRole)
	}

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer tx.Rollback()

	var userId int64
	err = tx.QueryRowContext(ctx, UPDATE_ROLE, newRole, username).Scan(&userId)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	err = db.AddNotifGeneric(tx, userId, adminId, notifType, notifMessage)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		go func(username string) {
			_, err := db.GetUserByUserName(nil, username, true)
			if err != nil {
				err = fmt.Errorf(errorMsg, err)
				log.Println(err.Error())
				return
			}
		}(username)
	}

	return nil
}

// ----------------------------------------
// GET FUNCS                               |
// ----------------------------------------

// Get user by id from User table
// If db.UseCache it places user on users cache
func (db *DataBase) GetUserById(tx *sql.Tx, id int64) (*config.User, error) {
	errorMsg := "GetUserById: %w"

	u := config.NewUser()

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	err := exor.QueryRowContext(ctx, GET_USER_BY_ID, id).
		Scan(
			&u.ID,
			&u.UserName,
			&u.ProfilePic,
			&u.Bio,
			&u.TotalKarma,
			&u.Created,
			&u.Role,
			&u.Banned,
			&u.BanExpDate,
			&u.BanReason,
			&u.BannedBy,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			u.ID = 0
			return u, fmt.Errorf(errorMsg, custom_errs.ErrIdNotFound)
		}
		return u, fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		db.putUserInCache(*u)
	}
	return u, nil
}

func (db *DataBase) GetUserBySub(sub string) (*config.User, error) {
	if sub == "" {
		return nil, errors.New("empty sub")
	}

	user := &config.User{}
	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, GET_USER_BY_SUB, sub).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.ProfilePic,
		&user.OAuthSub,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, custom_errs.ErrNameNotFound
		}
		return nil, fmt.Errorf("%v: %w", "database error", err)
	}

	return user, nil
}

func (db *DataBase) GetUserByCookie(sessionToken string) (*config.User, error) {
	errMsg := "get user by cookie: %w"
	u := config.NewUser()
	var expirationTime time.Time

	ctx, cancel := context.WithTimeout(db.Ctx, time.Second)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return u, fmt.Errorf(errMsg, err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, GET_USER_FROM_SESSION, sessionToken).
		Scan(
			&expirationTime,
			&u.ID,
			&u.UserName,
			&u.Email,
			&u.ProfilePic,
			&u.Bio,
			&u.TotalKarma,
			&u.Created,
			&u.Role,
			&u.Banned,
			&u.BanExpDate,
			&u.BanReason,
			&u.BannedBy,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return u, fmt.Errorf(errMsg, custom_errs.ErrNoRows)
		}
		return u, err
	}

	if time.Now().After(expirationTime) {
		result, err := tx.ExecContext(ctx, DELETE_SESSION, sessionToken)
		if err != nil {
			return u, fmt.Errorf(errMsg, err)
		}
		err = CatchNoRowsErr(result)
		return u, errors.New("session expired: " + err.Error())
	}

	if err = tx.Commit(); err != nil {
		return u, fmt.Errorf(errMsg, err)
	}

	if db.UseCache {
		db.putUserInCache(*u)
	}

	return u, nil
}

// Get user by username from User table
func (db *DataBase) GetUserByUserName(tx *sql.Tx, username string, getFromDb bool) (u *config.User, err error) {
	errorMsg := "GetUserByUserName: %w"

	if db.UseCache && !getFromDb {
		u, err := db.cache.Users.Get(username)
		if err == nil {
			log.Println("Found User In Cache")
			return u, nil
		}
	}

	u = config.NewUser()

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	err = exor.QueryRowContext(ctx, GET_USER_BY_NAME, username).
		Scan(
			&u.ID,
			&u.UserName,
			&u.ProfilePic,
			&u.Bio,
			&u.TotalKarma,
			&u.Created,
			&u.Role,
			&u.Banned,
			&u.BanExpDate,
			&u.BanReason,
			&u.BannedBy,
		)

	if err == sql.ErrNoRows {
		err = fmt.Errorf(errorMsg, custom_errs.ErrNameNotFound)
		return
	}

	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		db.putUserInCache(*u)
	}
	return
}

// Get total ammount of all users
func (db *DataBase) CountUsers() (int64, error) {
	if db.UseCache {
		if count, _ := db.cache.GetUserCount(); count != 0 {
			return count, nil
		}
	}

	var totalUsers int64

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, COUNT_USERS).Scan(&totalUsers)
	return totalUsers, err
}

func (db *DataBase) GetUserIdByPostId(postId int64) (int64, error) {
	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	var userId int64
	err := exor.QueryRowContext(ctx, GET_USER_ID_BY_POST_ID, postId).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("GET_USER_BY_POST_ID failed: %w", err)
	}
	return userId, nil
}

func (db *DataBase) UpdateUserKarma() error {

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	r, err := exor.ExecContext(ctx, UPDATE_USER_KARMA)
	if err != nil {
		return errors.New("failed to update users karma")
	}
	if err := CatchNoRowsErr(r); err != nil {
		return errors.New("no karma updated")
	}
	return nil
}
