package database

import (
	"fmt"
	"forum/common/custom_errs"
	"log"
	"time"
)

type Report struct {
	Id        int64
	SenderId  int64
	PostId    *int64
	CommentId *int64
	Message   string
	Created   time.Time
}

// To add report on a post call with commentId=0. Message can be an empty string
func (db *DataBase) Report(SenderId int64, postId int64, commentId int64, message string) error {
	var errorMsg = "report error: %w"
	var args []any
	args = append(args, SenderId)

	var query string
	if postId == 0 {
		query = fmt.Sprintf(ADD_REPORT, "CommentId")
		args = append(args, commentId)
	} else if commentId == 0 {
		query = fmt.Sprintf(ADD_REPORT, "PostId")
		args = append(args, postId)
	} else {
		return fmt.Errorf(errorMsg, custom_errs.ErrInvalidReportArgs)
	}
	args = append(args, message)

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	r, err := exor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	return nil
}

// To get all reports on a post call with commentId=0
func (db *DataBase) GetReports(commentId, postId int64) ([]*Report, error) {
	var reports []*Report
	var errorMsg = "get reports: %w"
	var query string
	var arg int64
	if postId == 0 {
		arg = commentId
		query = fmt.Sprintf(GET_REPORTS, "CommentId")
	} else if commentId == 0 {
		arg = postId
		query = fmt.Sprintf(GET_REPORTS, "PostId")
	} else {
		return nil, fmt.Errorf(errorMsg, custom_errs.ErrInvalidReportArgs)
	}

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()

	rows, err := exor.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf(errorMsg, err)
	}
	defer rows.Close()

	for rows.Next() {
		r := &Report{}
		err := rows.Scan(
			&r.Id, &r.SenderId, &r.PostId, &r.CommentId, &r.Message, &r.Created,
		)
		if err != nil {
			log.Println(errorMsg, err.Error())
			continue
		}

		reports = append(reports, r)
	}
	return reports, nil
}
