package database

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"time"
)

// TODO: Do we need to send notifs here?
// Removes or Restores/Approves content (user, post, comment)
func (db *DataBase) ModAssessment(table string, rowId int64, modId int64, reason, modName string, status int) error {
	var ok bool
	var actionType string
	var errorMsg = "moderator assessment error: %w"

	if _, ok = TABLES[table]; !ok {
		return fmt.Errorf(errorMsg, custom_errs.ErrInvalidTable)
	}

	if status == RESTORE {
		actionType = "approve"
	} else if status == REMOVE {
		actionType = "remove"
	} else {
		return fmt.Errorf(errorMsg, custom_errs.ErrInvalidDeleteStatus)
	}

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer tx.Rollback()

	r, err := tx.ExecContext(ctx, NEW_MOD_LOG, actionType, modId, table, rowId, reason)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	// var TargetParentId int64
	if _, err = db.ToggleRemove(tx, table, rowId, status, reason, modName); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		switch table {
		case "post":
			db.deleteOrRestorePostInCache(rowId, status, errorMsg)
		case "user":
			db.deleteOrRestoreUserInCache(rowId, status, errorMsg)
		}
	}

	return nil
}

func (db *DataBase) ToggleRemove(tx *sql.Tx, table string, rowId int64, status int, reason, modName string) (int64, error) {
	var errorMsg = "toggle remove: %w"
	if status != REMOVE && status != RESTORE {
		return 0, fmt.Errorf(errorMsg, custom_errs.ErrInvalidDeleteStatus)
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	toggle := func(q string) error {
		r, err := exor.ExecContext(ctx, q, status, reason, modName, time.Now(), rowId)
		if err != nil {
			return fmt.Errorf("toggle anon func: "+errorMsg, err)
		}
		return CatchNoRowsErr(r)
	}

	q := fmt.Sprintf(TOGGLE_REMOVE_STATUS, table)
	switch table {
	case "user":
		err := toggle(q)
		if err != nil {
			return 0, fmt.Errorf("toggle user: "+errorMsg, err)
		}
		err = db.toggleHideImage(tx, table, rowId, status)
		if err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}

	case "post":
		err := toggle(q)
		if err != nil {
			return 0, fmt.Errorf("toggle post: "+errorMsg, err)
		}

		err = db.toggleHideImage(tx, table, rowId, status)
		if err != nil {
			return 0, fmt.Errorf(errorMsg, err)
		}

	case "comment":
		q = q + " RETURNING UserID"
		var UserId int64
		err := tx.QueryRowContext(ctx, q, status, reason, modName, time.Now(), rowId).Scan(&UserId)
		if err != nil {
			return 0, fmt.Errorf("toggle comment: "+errorMsg, err)
		}

		return UserId, nil

	default:
		return 0, errors.New("toggle remove: error invalid table")
	}
	return 0, nil
}
