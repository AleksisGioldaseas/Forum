package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"time"
)

func CatchNoRowsErr(result sql.Result) error {
	row, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if row == 0 {
		return custom_errs.ErrIdNotFound
	}
	return nil
}

func reactionStringToInt(reaction string) (int, error) {
	switch reaction {
	case "neutral":
		return 0, nil
	case "like":
		return 1, nil
	case "dislike":
		return -1, nil
	default:
		return 0, custom_errs.ErrInvalidVoteAction
	}
}

func reactionIntToString(reaction int) (string, error) {
	switch reaction {
	case 0:
		return "neutral", nil
	case 1:
		return "like", nil
	case -1:
		return "dislike", nil
	default:
		return "", custom_errs.ErrInvalidVoteAction
	}
}

func tableToCol(table string) (string, error) {
	switch table {
	case "post":
		return "PostId", nil
	case "comment":
		return "CommentId", nil
	default:
		return "", errors.New("invalid table type")
	}
}

func (db *DataBase) checkIfContentRemoved(cntType string, cntId int64) error {
	errorMsg := "check if content removed: %w"
	if _, err := tableToCol(cntType); err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	var removed, deleted int
	q := fmt.Sprintf("SELECT Removed, Deleted FROM %s  WHERE Id = ?", cntType)
	err := db.conn.QueryRow(q, cntId).Scan(&removed, &deleted)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	if removed == 1 || deleted == 1 {
		return fmt.Errorf(errorMsg, custom_errs.ErrInteractionForbiden)
	}
	return nil
}

func returnImgColName(table string) (string, error) {
	var col string
	if table == "user" {
		col = "ProfilePic"
	} else if table == "post" {
		col = "Img"
	} else {
		return "", custom_errs.ErrInvalidTable
	}
	return col, nil
}

// it returns a context with timeout set to 1 second, and an executor, if given a tx, it will return it, if not it will return an executor from db.conn
func (db *DataBase) newCtxTx(tx *sql.Tx) (context.Context, context.CancelFunc, Executor) {

	ctx, cancel := context.WithTimeout(db.Ctx, time.Second)

	var conn Executor
	if tx != nil {
		conn = tx
	} else {
		conn = db.conn
	}
	return ctx, cancel, conn
}

func (db *DataBase) ExportTx() (*sql.Tx, error) {
	return db.conn.Begin()
}

func (db *DataBase) ExportTxContext(ctx context.Context) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, nil)
}
