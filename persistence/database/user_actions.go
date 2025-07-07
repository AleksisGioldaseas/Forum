package database

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"forum/server/core/sse"
	"log"
)

// Updates Deleted column in table status 1 for removed and 0 for active. Table UserId must be a match with activeUserId
func (db *DataBase) ToggleDeleteStatus(tx *sql.Tx, activeUserId int64, table string, rowId int64, status int) error {
	// var ok bool
	var errorMsg = "toggle delete status error: %w"
	if _, ok := TABLES[table]; !ok {
		return fmt.Errorf(errorMsg, custom_errs.ErrInvalidTable)
	}

	if status != REMOVE && status != RESTORE {
		return fmt.Errorf(errorMsg, custom_errs.ErrInvalidDeleteStatus)
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	toggle := func(col string) error {
		q := fmt.Sprintf(TOGGLE_DELETE_STATUS, table, col)
		r, err := exor.ExecContext(ctx, q, status, rowId, activeUserId)
		if err != nil {
			return fmt.Errorf(errorMsg, err)
		}

		if err = CatchNoRowsErr(r); err != nil {
			return fmt.Errorf(errorMsg, err)
		}
		return nil
	}

	switch table {
	case "user": // Currently not used !!
		err := toggle("Id")
		if err != nil {
			return fmt.Errorf(errorMsg, err)
		}

		err = db.toggleHideImage(tx, table, rowId, status)
		if err != nil {
			return fmt.Errorf(errorMsg, err)
		}

	case "post":
		err := toggle("UserId")
		if err != nil {
			return fmt.Errorf(errorMsg, err)
		}

	case "comment":
		return toggle("UserId")

	default:
		return errors.New("toggle delete status: error invalid table")
	}
	return nil
}

// Call with tx when uploading image or profile and image
func (db *DataBase) UpdateUserProfile(tx *sql.Tx, u *config.User) error {
	errorMsg := "UpdateUserProfile: %w"
	if u.Bio == nil && u.ProfilePic == nil {
		return fmt.Errorf(errorMsg, errors.New("no new values"))
	}

	var query string
	var args []any

	if u.Bio != nil && u.ProfilePic != nil {
		query = UPDATE_BIO_AND_PIC
		args = []any{u.Bio, u.ProfilePic, u.ID}
	} else if u.Bio != nil {
		query = UPDATE_BIO
		args = []any{u.Bio, u.ID}
	} else if u.ProfilePic != nil {
		query = UPDATE_PIC
		args = []any{u.ProfilePic, u.ID}
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	r, err := exor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	if err := CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if u.ProfilePic != nil {
		err = db.AddImage(tx, *u.ProfilePic)
		if err != nil {
			return fmt.Errorf(errorMsg, errors.New("add image failed"))
		}
	}

	if db.UseCache {
		// GetUserById updates cache so calling it here updates cache with the full user struct
		_, err = db.GetUserById(tx, u.ID)
		if err != nil {
			log.Println("UpdateUserProfile: error updating cache", err.Error())
		}
	}
	return nil
}

func (db *DataBase) ModeratorRequest(senderId int64, username string) error {

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	//TODO? leave notification for admins only that this user wants to be a moderator
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't add moderator request %w", err)
	}
	defer tx.Rollback()

	r, err := tx.ExecContext(ctx, `INSERT INTO ModRequest (SenderId) VALUES (?)`, senderId)
	if err != nil {
		return fmt.Errorf("insert into ModRequest table: %w", err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		return fmt.Errorf("insert into ModRequest table: %w", err)
	}

	adminIds, err := db.getAdminIds(tx)
	if err != nil {
		return fmt.Errorf("issue with getting admin id's: %w", err)
	}

	for _, adminId := range adminIds {
		err = db.AddNotifGeneric(tx, adminId, senderId, "modrequest", username+" has requested to become a mod")
		if err != nil {
			return fmt.Errorf("failed to add notification to admin: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit moderator request notifications: %w", err)
	}

	for _, adminId := range adminIds {
		err := sse.SendSSENotification(adminId)
		if err != nil {
			if errors.Is(err, custom_errs.ErrUserNotConnected) {
				log.Println("Moderator Request: SendSSENotification failed:", err.Error())
			} else {
				return err
			}
		}
	}

	return nil
}
