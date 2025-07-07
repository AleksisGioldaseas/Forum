package database

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/config"
	"forum/server/core/sse"
	"strings"
)

func (db *DataBase) AddComment(c *config.Comment) (int64, error) {
	errorMsg := "add comment: %v"
	if c.Body == "" {
		return 0, fmt.Errorf(errorMsg, "empty comment")
	}

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	defer tx.Rollback()

	if err := db.checkIfContentRemoved("post", c.PostID); err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	result, err := tx.ExecContext(
		ctx,
		CREATE_COMMENT,
		c.UserID,
		c.PostID,
		c.Body,
	)

	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	commentId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	var postTitle string
	err = tx.QueryRowContext(ctx, "SELECT Title FROM Post WHERE Id = ?", c.PostID).Scan(&postTitle)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	err = db.AddUserActivity(tx, c.UserID, nil, &commentId, nil)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	notifReceiverId, err := db.AddNotifComment(tx, c.PostID, commentId)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	err = sse.SendSSENotification(notifReceiverId)
	if err != nil {
		if errors.Is(err, custom_errs.ErrUserNotConnected) {
		} else {
			return commentId, errors.Join(custom_errs.ErrNotificationFailed, err)
		}
	}

	return commentId, nil
}

// Need to have comment Id and new body
func (db *DataBase) EditComment(comId int64, body string, activeUserId int64) error {

	errorMsg := "edit comment: %w"

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	result, err := exor.ExecContext(ctx, UPDATE_COMMENT, body, comId, activeUserId)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	err = CatchNoRowsErr(result)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	return nil
}

// Returns ErrIdNotFound error if id is not present in table
func (db *DataBase) DeleteCom(activeUserId, comId int64) error {
	errMsg := "delete comment: %w"

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	defer tx.Rollback()

	if err := db.ToggleDeleteStatus(tx, activeUserId, "comment", comId, REMOVE); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	if err = db.DeleteUserActivity(
		tx,
		activeUserId, nil, &comId, nil,
	); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	return nil
}

// Check len(comments) for any results. Look through []errors for any errors
// - Limit is the number of posts per page
// - Offset start from 0
func (db *DataBase) GetComments(tx *sql.Tx, postId int, limit int, offset int, activeUserId int64, activeUserRole int, orderBy string, reportedOnly, removedOnly bool) (comments []*config.Comment, errors []error) {
	errorMsg := "GetComByPostId: %w"
	if limit > db.Limits.RowsLimit {
		return nil, []error{fmt.Errorf(errorMsg, custom_errs.ErrExceededRowsLimit)}
	}

	args := []any{activeUserRole, activeUserRole, activeUserRole, activeUserRole, activeUserRole, activeUserId, postId, removedOnly, reportedOnly, limit, offset * limit}

	switch orderBy {
	case "new":
		orderBy = "c.Created DESC"
	case "old":
		orderBy = "c.Created ASC"
	case "top":
		orderBy = "c.TotalKarma DESC"
	}

	extra := ""
	if removedOnly {
		orderBy = "c.RemovedTime DESC"
		extra = " AND c.Removed = 1"
	}

	if reportedOnly {
		extra = " AND r.CommentId IS NOT NULL"
		orderBy = "MAX(r.Created) DESC"
	}

	q := fmt.Sprintf(GET_COMMENTS, extra, orderBy)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	rows, err := exor.QueryContext(ctx, q, args...)
	if err != nil {
		errors = []error{fmt.Errorf(errorMsg, err)}
		return
	}
	defer rows.Close()

	for rows.Next() {
		c := config.NewComment()
		var reports string
		err := rows.Scan(
			&c.ID, &c.UserID, &c.PostID,
			&c.Body, &c.CreationDate, &c.UserName, &c.Removed, &c.Deleted, &c.Edited, &c.UserRole,
			&c.Likes, &c.Dislikes, &c.TotalKarma, &c.UserReaction, &reports, &c.RemovalReason, &c.ModeratorName,
		)

		if err != nil {
			errors = append(errors, fmt.Errorf(errorMsg, err))
			continue
		}

		c.Reports = strings.Split(reports, "|!|!|")

		comments = append(comments, c)
	}

	return
}

func (db *DataBase) CountComments(postId int64) (int, error) {
	errorMsg := "CountComments: %w"

	var totalComments int

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, COUNT_COMMENTS, postId).Scan(&totalComments)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	return totalComments, nil

}

func (db *DataBase) GetCommentById(tx *sql.Tx, activeUserRole int, commentId int64) (*config.Comment, error) {
	errorMsg := "GetComByComId: %w"

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	rows, err := exor.QueryContext(ctx, GET_COMMENT, activeUserRole, activeUserRole, activeUserRole, commentId)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}
	defer rows.Close()

	c := config.NewComment()

	if rows.Next() {
		var reports string

		err := rows.Scan(
			&c.ID, &c.UserID, &c.PostID,
			&c.Body, &c.CreationDate, &c.UserName, &c.Removed, &c.Deleted, &c.UserRole,
			&c.Likes, &c.Dislikes, &c.TotalKarma, &reports, &c.RemovalReason, &c.ModeratorName,
		)
		if err != nil {
			return nil, fmt.Errorf(errorMsg, err)
		}

		c.Reports = strings.Split(reports, "|!|!|")

	} else {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf(errorMsg, err)
		}
		return nil, fmt.Errorf(errorMsg, sql.ErrNoRows)
	}

	return c, nil
}
