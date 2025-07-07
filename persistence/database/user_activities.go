package database

import (
	"database/sql"
	"fmt"
	"forum/common/custom_errs"
	"time"
)

type UserActivity struct {
	Id         int64         `json:"id"`         // the row id
	UserId     int64         `json:"-"`          // the acting user
	ActionType string        `json:"actionType"` // post, comment, like, dislike
	ReactionId sql.NullInt64 `json:"reactionId"` // reaction row id used to fetch the rest of the data
	PostId     sql.NullInt64 `json:"postId"`     // if activity is post or reaction on post
	CommentId  sql.NullInt64 `json:"commentId"`  // if activity is comment or reaction on comment
	Created    time.Time     `json:"created"`    // time the activity was logged
	Preview    string        `json:"preview"`    // title or snippet of content
	BonusText  string        `json:"bonusText"`  // extra snippet (e.g. post or comment body)
}

func (db *DataBase) AddUserActivity(tx *sql.Tx, userId int64, postId, commentId, reactionId *int64) error {
	errorMsg := "add user activity: %w"

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	r, err := exor.ExecContext(ctx, INSERT_USER_ACTIVITY,
		userId, postId, commentId, reactionId)

	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	err = CatchNoRowsErr(r)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	return nil
}

func (db *DataBase) DeleteUserActivity(
	tx *sql.Tx,
	userId int64, postId, commentId, reactionId *int64,

) error {
	var targetId string
	var value *int64
	errorMsg := "delete user activity: %w"

	if postId != nil {
		targetId = "postId"
		value = postId
	} else if commentId != nil {
		targetId = "commentID"
		value = commentId
	} else if reactionId != nil {
		targetId = "reactionId"
		value = reactionId
	}

	q := fmt.Sprintf(DELETE_USER_ACTIVITY, targetId)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	r, err := exor.ExecContext(ctx, q, userId, value)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	err = CatchNoRowsErr(r)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	return nil
}

func (db *DataBase) GetUserActivitiesByUname(tx *sql.Tx, userName string, limit, page int) ([]*UserActivity, error) {
	errMsg := "get user activities: %w"
	if limit > db.Limits.RowsLimit {
		return nil, fmt.Errorf(errMsg, custom_errs.ErrExceededRowsLimit)
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	rows, err := exor.QueryContext(ctx, GET_USER_ACTIVITIES_BY_UNAME, userName, limit, limit*page)
	if err != nil {
		return nil, fmt.Errorf("GET_ACTIVITIES_BY_USERID query failed: %w", err)
	}
	defer rows.Close()

	var userActivities []*UserActivity

	for rows.Next() {
		u := &UserActivity{}
		err := rows.Scan(&u.Id, &u.UserId, &u.PostId, &u.CommentId, &u.ReactionId, &u.Created)
		if err != nil {
			return nil, fmt.Errorf("scanning GET_ACTIVITIES_BY_USERID rows failed: %w", err)
		}

		if err = db.parseUserActivity(u); err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}

		userActivities = append(userActivities, u)
	}

	return userActivities, nil
}

// Depracated
func (db *DataBase) GetUserActivities(tx *sql.Tx, userId int64, limit, offset int) ([]*UserActivity, error) {
	errMsg := "get user activities: %w"
	if limit > db.Limits.RowsLimit {
		return nil, fmt.Errorf(errMsg, custom_errs.ErrExceededRowsLimit)
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	rows, err := exor.QueryContext(ctx, GET_USER_ACTIVITIES, userId, limit, limit*offset)
	if err != nil {
		return nil, fmt.Errorf("GET_ACTIVITIES_BY_USERID query failed: %w", err)
	}
	defer rows.Close()

	var userActivities []*UserActivity

	for rows.Next() {
		u := &UserActivity{}
		err := rows.Scan(&u.Id, &u.UserId, &u.PostId, &u.CommentId, &u.ReactionId, &u.Created)
		if err != nil {
			return nil, fmt.Errorf("scanning GET_ACTIVITIES_BY_USERID rows failed: %w", err)
		}

		if err = db.parseUserActivity(u); err != nil {
			return nil, fmt.Errorf(errMsg, err)
		}

		userActivities = append(userActivities, u)
	}

	return userActivities, nil
}

func (db *DataBase) parseUserActivity(ua *UserActivity) error {
	var err error
	switch {
	case ua.PostId.Valid:
		err = db.parsePost(ua)
	case ua.CommentId.Valid:
		err = db.parseComment(ua)
	case ua.ReactionId.Valid:
		err = db.parseReaction(ua)
	}
	return err
}

func (db *DataBase) parsePost(ua *UserActivity) error {
	errMsg := "parse post %w"
	var removed, deleted int
	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, GET_POST_ACTIVITY, ua.PostId.Int64).
		Scan(
			&ua.Preview, &ua.BonusText, &removed, &deleted,
		)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	ua.ActionType = "post"

	if removed == TRUE || deleted == TRUE {
		ua.PostId.Valid = false
		ua.Preview = REMOVED_OR_DELETED_CONTENT
		ua.BonusText = ""
	}
	return err
}

func (db *DataBase) parseComment(ua *UserActivity) error {
	errMsg := "parse comment %w"
	var removed, deleted int
	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	err := exor.QueryRowContext(ctx, GET_COM_ACTIVITY, ua.CommentId.Int64).
		Scan(
			&ua.PostId, &ua.Preview, &ua.BonusText, &removed, &deleted,
		)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	ua.ActionType = "comment"

	if removed == 1 || deleted == 1 {
		ua.PostId.Valid = false
		ua.Preview = REMOVED_OR_DELETED_CONTENT
	}
	return nil
}

func (db *DataBase) parseReaction(ua *UserActivity) error {
	errMsg := "parse reaction: %w"
	var Reaction int

	ctx, cancel, exor := db.newCtxTx(nil)
	defer cancel()
	if err := exor.QueryRowContext(ctx, GET_USER_REACTION, ua.ReactionId.Int64).Scan(
		&ua.UserId, &ua.PostId, &ua.CommentId, &Reaction,
	); err != nil {
		return fmt.Errorf(errMsg, err)
	}

	ua.ActionType, _ = reactionIntToString(Reaction)

	var removed, deleted int
	if ua.PostId.Valid {
		err := exor.QueryRowContext(ctx, GET_POST_ACTIVITY, ua.PostId.Int64).
			Scan(
				&ua.Preview, &ua.BonusText, &removed, &deleted,
			)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}
		if removed == TRUE || deleted == TRUE {
			ua.PostId.Valid = false
			ua.Preview = REMOVED_OR_DELETED_CONTENT
			ua.BonusText = ""
		}
	} else if ua.CommentId.Valid {
		err := exor.QueryRowContext(ctx, GET_COM_ACTIVITY, ua.CommentId.Int64).Scan(
			&ua.PostId, &ua.Preview, &ua.BonusText, &removed, &deleted,
		)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}
	}
	return nil
}
