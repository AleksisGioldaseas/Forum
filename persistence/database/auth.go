package database

/* Checks for UNIQUE username and email*/

import (
	"database/sql"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"net/http"
	"time"
)

func (db *DataBase) IsUsernameUnique(tx *sql.Tx, username string) (bool, error) {
	var exists bool

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	err := exor.QueryRowContext(ctx, CHECK_USERNAME_UNIQUE, username).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		// Other database error
		return false, err
	}
	return false, nil
}

func (db *DataBase) IsEmailUnique(tx *sql.Tx, username string) (bool, error) {
	var exists bool

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	err := exor.QueryRowContext(ctx, CHECK_EMAIL_UNIQUE, username).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		// Other database error
		return false, err
	}
	return false, nil
}

func (db *DataBase) GetAuthUserByUserName(username string) (*config.User, error) {

	u := config.NewUser()

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, GET_USER_INFO_BY_UNAME, username).
		Scan(
			&u.ID,
			&u.UserName,
			&u.Email,
			&u.PasswordHash,
			&u.Salt,
		)

	if err == sql.ErrNoRows {
		return nil, custom_errs.ErrNameNotFound
	}
	return u, nil
}

func (db *DataBase) UpdatePassword(userId, newPass string) (err error) {

	if newPass == "" {
		return custom_errs.ErrInvalidPassword
	}

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	result, err := exor.ExecContext(ctx, UPDATE_PASS, newPass, userId)
	if err != nil {
		return
	}

	row, err := result.RowsAffected()
	if row == 0 {
		err = custom_errs.ErrNoRows
	}
	return
}

func (db *DataBase) UpdateUserData(u *config.User) (err error) {

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	result, err := exor.ExecContext(ctx, UPDATE_USER_DATA, u.UserName, u.Email, u.ID)
	if err != nil {
		return
	}

	if err := CatchNoRowsErr(result); err != nil {
		return err
	}
	return nil
}

func (db *DataBase) StoreSession(sessionToken string, uID int64, expiresAt time.Time) error {

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	result, err := exor.ExecContext(ctx, STORE_SESSION, sessionToken, uID, expiresAt)
	if err != nil {
		return err
	}
	err = CatchNoRowsErr(result)
	return err
}

func (db *DataBase) GetOtherSessions(uID int64, currentSessionToken string) ([]string, error) {

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	rows, err := exor.QueryContext(ctx,
		GET_OTHER_SESSIONS,
		uID, currentSessionToken, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []string
	for rows.Next() {
		var session string
		if err := rows.Scan(&session); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (db *DataBase) DeleteSessions(sessionTokens []string) error {

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, token := range sessionTokens {
		_, err := tx.ExecContext(ctx, DELETE_SESSION, token)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DataBase) DeleteSession(sessionToken string) error {

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	result, err := exor.ExecContext(ctx, DELETE_SESSION, sessionToken)
	if err != nil {
		return err
	}
	err = CatchNoRowsErr(result)
	return err
}

// Checks if the user already has an active session open
func (db *DataBase) HasSession(r *http.Request) (bool, error) {

	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return false, nil
		}
		return false, fmt.Errorf("failed to read cookie: %w", err)
	}

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	var count int
	err = exor.QueryRowContext(ctx, HAS_SESSION,
		cookie.Value, time.Now(),
	).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("database error: %w", err)
	}

	return count > 0, nil
}
