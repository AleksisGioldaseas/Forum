package handlers

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func VoteCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Vote comment handler called!")

	type CommentVoteRequest struct {
		Id     int64  `json:"id"`
		Action string `json:"action"` //"like", "dislike", "neutral"
	}

	vote := &CommentVoteRequest{}
	err := jsonRequestExtractor(request, vote)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "vote comment: "+err.Error())
		return
	}

	err, _ = db.React(int64(vote.Id), activeUser.ID, "comment", vote.Action)
	if err != nil && !errors.Is(err, custom_errs.ErrUserNotConnected) {
		fmt.Println("Database update failed:", err)
		if err == custom_errs.ErrDuplicateReaction {
			jsonProblemResponder(writer, http.StatusOK, "You did that already", "VoteComment: Failed to update comment likes/dislikes. "+err.Error())
			return
		}
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "VoteComment: Failed to update comment likes/dislikes. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("VoteComment: Failed to send final response!")
	}
}
