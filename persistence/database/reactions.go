package database

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/server/core/sse"
	"log"
)

// contentTypeCol = (PostId, CommentId)
// table = (Post, Comment)

func (db *DataBase) React(
	contentId, activeUserId int64,
	table string,
	newReactionString string,
) (error, bool) {
	var errorMsg = "react: %w"
	if err := db.checkIfContentRemoved(table, contentId); err != nil {
		return fmt.Errorf(errorMsg, err), false
	}

	var prevReaction int
	var firstReaction bool
	var bonusText string
	var notifReceiverId int64 // This is Zero if the user reacts on their own content

	newReactionInt, err := reactionStringToInt(newReactionString)
	if err != nil {
		return fmt.Errorf(errorMsg, err), false
	}

	karmaValue := newReactionInt

	contentTypeCol, err := tableToCol(table)
	if err != nil {
		return fmt.Errorf(errorMsg, err), false
	}

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err, false
	}
	defer tx.Rollback()

	if prevReaction, firstReaction, err = db.getReaction(
		tx,
		activeUserId, contentId,
		contentTypeCol,
	); err != nil {
		return fmt.Errorf(errorMsg, err), false
	}

	if newReactionInt == prevReaction {
		return fmt.Errorf(errorMsg, custom_errs.ErrDuplicateReaction), false
	}

	var reactionId int64

	if firstReaction {
		if reactionId, err = db.makeNewReaction(
			tx,
			contentId, activeUserId,
			contentTypeCol, table,
			&bonusText, newReactionInt,
		); err != nil {
			return fmt.Errorf(errorMsg, err), false
		}

		err = db.AddUserActivity(tx,
			activeUserId, nil, nil, &reactionId)

		if err != nil {
			return fmt.Errorf(errorMsg, err), firstReaction
		}
		if notifReceiverId, err = db.AddNotifReaction(tx, reactionId); err != nil {
			return fmt.Errorf(errorMsg, err), false
		}
	} else {
		switch newReactionInt {
		case 0:
			karmaValue = -prevReaction
			var reactionId int64
			if reactionId, err = db.deleteReaction(tx,
				contentId, activeUserId,
				table, prevReaction, contentTypeCol,
			); err != nil {
				return fmt.Errorf(errorMsg, err), false
			}

			if err = db.DeleteUserActivity(tx,
				activeUserId, nil, nil, &reactionId,
			); err != nil {
				return fmt.Errorf(errorMsg, err), false
			}

		case 1, -1:
			if newReactionInt == -prevReaction {
				karmaValue += karmaValue

				if err = db.toggleReaction(tx,
					contentId, activeUserId,
					table, contentTypeCol,
					newReactionInt,
				); err != nil {
					return fmt.Errorf(errorMsg, err), false
				}
			}
		}
	}

	err = db.updateUserKarma(tx, contentId, table, karmaValue)
	if err != nil {
		return fmt.Errorf(errorMsg, err), false
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf(errorMsg, err), false
	}

	var notifErr error
	if firstReaction {
		notifErr = sse.SendSSENotification(notifReceiverId)
		if notifErr != nil && errors.Is(err, custom_errs.ErrUserNotConnected) {
			if errors.Is(notifErr, custom_errs.ErrUserNotConnected) {
				log.Println("react: SendSSENotification failed:", notifErr.Error())
			} else {
				notifErr = errors.Join(custom_errs.ErrNotificationFailed, err)
			}
		}
	}

	if table == "post" && db.UseCache {
		go func() {
			p, err := db.GetPostById(nil, -1, contentId, false)
			if err == nil {
				db.putPostInCache(*p)
			} else {
				wrappedErr := fmt.Errorf(errorMsg, err)
				fmt.Println(wrappedErr)
			}

			u, err := db.GetUserById(nil, p.UserID)
			if err == nil {
				db.putUserInCache(*u)
			} else {
				wrappedErr := fmt.Errorf(errorMsg, err)
				fmt.Println(wrappedErr)
			}
		}()
	}

	// returning possible ErrNotificationFailed from notification.
	// This should be ignored by the handler
	return notifErr, firstReaction
}

func (db *DataBase) deleteReaction(
	tx *sql.Tx,
	contentId, activeUserId int64,
	table string,
	prevReaction int,
	contentTypeCol string,
) (reactionId int64, err error) {

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()

	errorMsg := "delete reaction: %w"
	query := fmt.Sprintf(DELETE_REACTION, contentTypeCol)
	err = exor.QueryRowContext(ctx, query, contentId, activeUserId).Scan(&reactionId)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	if err = db.undoReaction(tx, contentId, table, prevReaction); err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}
	return reactionId, nil
}

func (db *DataBase) undoReaction(
	tx *sql.Tx,
	contentId int64,
	table string,
	prevReaction int,
) error {

	errorMsg := "undo reaction: %w"
	var colToEdit string // like or dislike
	switch prevReaction {
	case 1:
		colToEdit = "likes"
	case -1:
		colToEdit = "dislikes"
	default:
		return fmt.Errorf(errorMsg, "invalid previews reaction")
	}

	query := fmt.Sprintf(UNDO_REACTION, table, colToEdit, colToEdit)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	r, err := exor.ExecContext(ctx, query, contentId)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	} else if err := CatchNoRowsErr(r); err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	return nil
}

func (db *DataBase) makeNewReaction(
	tx *sql.Tx,
	contentId, activeUserId int64,
	contentTypeCol, table string,
	bonusText *string,
	reaction int,
) (reactionId int64, err error) {

	errorMsg := "new reaction: %w"
	var query string

	// Updating reactions
	query = fmt.Sprintf(NEW_USER_REACTIONS_GEN, contentTypeCol)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	result, err := exor.ExecContext(ctx, query, reaction, contentId, activeUserId)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	reactionId, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	// Updating Content and returning preview for notification
	var textRow, setRow string

	switch table {
	case "post":
		textRow = "Title"
	case "comment":
		textRow = "Body"
	}

	switch reaction {
	case 1:
		setRow = "Likes"
	case -1:
		setRow = "Dislikes"
	}

	query = fmt.Sprintf(UPDATE_REACTION_ON_CNT_AND_RETURN_TXT,
		table, setRow, setRow, textRow)

	err = exor.QueryRowContext(ctx, query, contentId).Scan(bonusText)
	if err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	} else if err := CatchNoRowsErr(result); err != nil {
		return 0, fmt.Errorf(errorMsg, err)
	}

	if len(*bonusText) > 40 {
		*bonusText = (*bonusText)[:40]
	}
	return reactionId, nil
}

func (db *DataBase) toggleReaction(
	tx *sql.Tx,
	contentId, activeUserId int64,
	table, col string,
	newReaction int,
) error {
	errorMsg := "toggle reaction: %w"
	query := fmt.Sprintf(TOGGLE_USER_REACTIONS, col)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	result, err := exor.ExecContext(ctx, query, contentId, activeUserId)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	if err := CatchNoRowsErr(result); err != nil {
		return err
	}
	query = fmt.Sprintf(TOGGLE_REACTION_ON_CNT, table)

	result, err = exor.ExecContext(
		ctx, query,
		newReaction, newReaction,
		contentId)

	if err != nil {
		return fmt.Errorf(errorMsg, err)
	} else if err := CatchNoRowsErr(result); err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	return nil
}

func (db *DataBase) updateUserKarma(
	tx *sql.Tx,
	contentId int64,
	contentTypeCol string,
	karmaValue int,
) error {

	errorMsg := "update user karma: %w"
	query := fmt.Sprintf(UPDATE_USER_KARMA_GEN, contentTypeCol)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	result, err := exor.ExecContext(ctx, query, karmaValue, contentId)
	if err != nil {
		return fmt.Errorf(errorMsg, custom_errs.ErrUpdatingUserKarma)
	} else if err := CatchNoRowsErr(result); err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	return nil
}

// Looks into UserReactions and returns old reaction code and
// firstReaction bool = false if exists or 0 reaction code and true
func (db *DataBase) getReaction(
	tx *sql.Tx,
	userId,
	contentId int64,
	contentTypeCol string,
) (int, bool, error) {

	var firstReaction bool

	// TODO check for valid content types
	query := fmt.Sprintf(GET_FROM_USER_REACTIONS_GEN, contentTypeCol)

	ctx, cancel, exor := db.newCtxTx(tx)
	defer cancel()
	row := exor.QueryRowContext(ctx, query, contentId, userId)
	var reaction int
	err := row.Scan(&reaction)
	if err != nil {
		if err == sql.ErrNoRows {
			reaction = 0
			firstReaction = true
			return reaction, firstReaction, nil
		}
		return 0, firstReaction, err
	}

	if reaction > 1 || reaction < -1 {
		return reaction, firstReaction, custom_errs.ErrUnknownReactionInDb
	}

	return reaction, firstReaction, nil
}
