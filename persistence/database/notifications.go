package database

import (
	"context"
	"database/sql"
	"fmt"
	"forum/common/custom_errs"
	"log"
	"time"
)

// Type corresponds to the table and TypeId to the row Id
// Seen 0, 1 is a bool value.
type Notification struct {
	Id             int64  //notificaiton id
	UserId         int64  `json:"-"` //user id that received notification
	SenderUserName string //who did the action that the notification is about
	ActionType     string // like, dislike, comment, mod action, mod request,ban,unban
	TargetType     string //post, comment, user
	TargetId       int64  //the id of the THING
	TargetParentId *int64 //the id of the parent of the thing, atm it will ever only be the postid if the target is a comment
	Seen           bool   //if this is a new notification or it has been already seen
	Created        time.Time
	BonusText      string
}

type NotificationResult struct {
	ReceiverId          int64
	SenderId            int64
	Seen                bool
	NotificationCreated time.Time
	NotifType           sql.NullString

	SenderUserName sql.NullString

	SuperPostTitle sql.NullString
	SuperPostId    sql.NullInt64

	Reaction sql.NullInt64

	ReactedCommentId        sql.NullInt64
	ReactedCommentPostTitle sql.NullString
	ReactedCommentPostId    sql.NullInt64

	PostReactedTitle sql.NullString
	PostReactedId    sql.NullInt64

	// Comment on post
	CommentOnPostId   sql.NullInt64
	CommentOnPostBody sql.NullString

	// Post that was commented
	PostCommentedTitle   sql.NullString
	PostCommentedId      sql.NullInt64
	PostCommentedIsSuper sql.NullBool

	// Bonus text for generic notifs
	BonusText sql.NullString
}

var notifTypes = map[string]struct{}{
	"user-promotion": {},
	"mod-demotion":   {},
	"modrequest":     {},
	"mod-promotion":  {},
	"ban":            {},
	"unban":          {},
}

func (db *DataBase) GetNotifications(tx *sql.Tx, receiverId int64, limit, offset int) (
	hasUnseen bool, notifications []*Notification, countAll, countUnseen int, err error) {
	errMsg := "get notifs: %w"
	if limit > db.Limits.RowsLimit {
		return false, nil, 0, 0, fmt.Errorf(errMsg, custom_errs.ErrExceededRowsLimit)
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	rows, err := exor.QueryContext(ctx, GET_NOTIFS, receiverId, limit, offset*limit)
	if err != nil {
		return false, nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var n NotificationResult
		err := rows.Scan(
			// Not Nulls
			&n.ReceiverId,
			&n.SenderId,
			&n.SenderUserName,
			&n.Seen,
			&n.NotificationCreated,

			// Null on comment, post, reactions
			&n.NotifType,

			// Super Post
			&n.SuperPostId,
			&n.SuperPostTitle,

			// Reactions
			&n.Reaction,

			// Comment reacted
			&n.ReactedCommentId,
			&n.ReactedCommentPostId,
			&n.ReactedCommentPostTitle,

			// Post reacted
			&n.PostReactedId,
			&n.PostReactedTitle,

			// Comment on post
			&n.CommentOnPostId,
			&n.CommentOnPostBody,

			// Post that was commented
			&n.PostCommentedId,
			&n.PostCommentedTitle,
			&n.PostCommentedIsSuper,

			&n.BonusText,

			// Total count
			&countAll,
		)
		if err != nil {
			return false, nil, 0, 0, err
		}

		if n.SenderId == n.ReceiverId {
			continue
		}

		parsed, err := parseNotif(n)
		if err != nil {
			log.Printf("get notifs: %v\n", err)
			continue
			// return nil, 0, fmt.Errorf("get notifs: %w", err)
		}
		if parsed == nil {
			continue
		}
		if !parsed.Seen {
			hasUnseen = true
			countUnseen++
		}
		notifications = append(notifications, parsed)
	}

	return hasUnseen, notifications, countAll, countUnseen, nil
}

func parseNotif(r NotificationResult) (*Notification, error) {
	n := &Notification{}

	n.UserId = r.ReceiverId
	n.Seen = r.Seen
	n.Created = r.NotificationCreated
	n.SenderUserName = r.SenderUserName.String

	var err error
	switch {
	case r.CommentOnPostId.Valid:
		err = parsePostComment(r, n)
	case r.PostReactedId.Valid:
		err = parsePostreaction(r, n)
	case r.Reaction.Valid && r.ReactedCommentId.Valid:
		err = parseCommentReaction(r, n)
	case r.SuperPostId.Valid:
		err = parseSuperReport(r, n)
	case r.NotifType.Valid:
		err = parseGeneric(r, n)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("parse notif: %w", err)
	}
	return n, nil
}

func parsePostComment(r NotificationResult, n *Notification) error {
	if !r.PostCommentedId.Valid || !r.CommentOnPostId.Valid || !r.PostCommentedTitle.Valid {
		return fmt.Errorf("parse post comment %w", custom_errs.ErrNullValueOnStructField)
	}
	n.ActionType = "comment"
	n.TargetId = r.CommentOnPostId.Int64
	n.TargetParentId = &r.PostCommentedId.Int64
	if !r.PostCommentedIsSuper.Bool {
		n.TargetType = "post"
	} else {
		n.TargetType = "super-report"
	}
	n.BonusText = r.PostCommentedTitle.String
	return nil
}

func parsePostreaction(r NotificationResult, n *Notification) error {
	if !r.Reaction.Valid || !r.PostReactedId.Valid || !r.PostReactedTitle.Valid {
		return fmt.Errorf("parse post reaction %w", custom_errs.ErrNullValueOnStructField)
	}
	n.TargetType = "post"
	n.ActionType, _ = reactionIntToString(int(r.Reaction.Int64))
	n.TargetId = r.PostReactedId.Int64
	n.BonusText = r.PostReactedTitle.String
	return nil
}

func parseCommentReaction(r NotificationResult, n *Notification) error {
	if !r.Reaction.Valid || !r.ReactedCommentId.Valid ||
		!r.ReactedCommentPostTitle.Valid || !r.ReactedCommentPostId.Valid {
		return fmt.Errorf("parse comment reaction %w", custom_errs.ErrNullValueOnStructField)
	}
	n.TargetType = "comment"
	n.ActionType, _ = reactionIntToString(int(r.Reaction.Int64))
	n.TargetId = r.ReactedCommentId.Int64
	n.BonusText = r.ReactedCommentPostTitle.String
	n.TargetParentId = &r.ReactedCommentPostId.Int64
	return nil
}

func parseSuperReport(r NotificationResult, n *Notification) error {
	if !r.SuperPostTitle.Valid || !r.SuperPostId.Valid {
		return fmt.Errorf("parse super report: %w", custom_errs.ErrNullValueOnStructField)
	}
	n.TargetType = "super-report"
	n.ActionType = "super-report"
	n.TargetId = r.SuperPostId.Int64
	n.BonusText = r.SuperPostTitle.String
	return nil
}

func parseGeneric(r NotificationResult, n *Notification) error {
	if _, ok := notifTypes[r.NotifType.String]; ok {
		n.ActionType = r.NotifType.String

		if !r.SenderUserName.Valid {
			n.SenderUserName = "System Administration"
		}

		if r.BonusText.Valid {
			n.BonusText = r.BonusText.String
		}
	} else {
		return fmt.Errorf("invalid notification type %s", r.NotifType.String)
	}
	return nil
}

func (db *DataBase) AddNotifSuper(tx *sql.Tx, receiverId, senderId, postId int64) error {
	if receiverId == senderId {
		return nil
	}
	errMsg := "add super notif: %w"

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	r, err := exor.ExecContext(ctx, ADD_NOTIF_SUPER, receiverId, senderId, postId)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (db *DataBase) AddNotifComment(tx *sql.Tx, postId, commentId int64) (receiverId int64, err error) {
	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	err = exor.QueryRowContext(ctx, ADD_NOTIF_COMMENT, postId, commentId).Scan(&receiverId)
	if err != nil {
		// Returning nil to continue the tx
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("notif comment: %w", err)
	}
	return receiverId, nil
}

func (db *DataBase) AddNotifReaction(tx *sql.Tx, UserReactionId int64) (receiverId int64, err error) {
	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	err = exor.QueryRowContext(ctx, ADD_NOTIF_REACTION, UserReactionId).Scan(&receiverId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return receiverId, nil
}

func (db *DataBase) AddNotifGeneric(tx *sql.Tx, receiverId, senderId int64, notifType, bonusText string) error {
	errMsg := "add notif plain: %w"
	if _, ok := notifTypes[notifType]; !ok {
		return fmt.Errorf(errMsg, "invalid notif type")
	}

	// Returning nil to continue the tx
	if senderId == receiverId {
		return nil
	}

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	r, err := exor.ExecContext(ctx, INSERT_NOTIF, receiverId, senderId, notifType, bonusText)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf(errMsg, err)
	}
	if err = CatchNoRowsErr(r); err != nil {
		fmt.Println(err)
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

func (db *DataBase) AllNotificationsSeen(tx *sql.Tx, userId int64) error {
	errMsg := "all notifications seen %w"

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	var result sql.Result
	result, err := exor.ExecContext(ctx, MARK_ALL_SEEN, userId)
	if err != nil {
		return fmt.Errorf(errMsg+" mark all seen", err)
	}
	if err := CatchNoRowsErr(result); err != nil {
		return fmt.Errorf(errMsg+" mark all seen", err)
	}
	// TODO: Depracated?
	result, err = exor.ExecContext(ctx, SEEN_FALSE, userId)
	if err != nil {
		return fmt.Errorf(errMsg+" seen false", err)
	}
	if err := CatchNoRowsErr(result); err != nil {
		return fmt.Errorf(errMsg+" seen false", err)
	}

	return nil
}

func (db *DataBase) CountNotifications(parentCtx *context.Context, tx *sql.Tx, userId int) (bool, int, error) {

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	if parentCtx != nil {
		ctx = *parentCtx
	}

	var totalNotifs int

	if err := exor.QueryRowContext(ctx, COUNT_NOTIFS, userId).Scan(&totalNotifs); err != nil {
		return false, 0, fmt.Errorf("count notifications: %w", err)
	}
	var hasUnseen bool
	if totalNotifs > 0 {
		hasUnseen = true
	}

	return hasUnseen, totalNotifs, nil
}
