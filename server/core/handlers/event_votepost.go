package handlers

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

const (
	LIKE    = 1
	DISLIKE = -1
	NEUTRAL = 0
)

func VotePostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Vote post handler called!")

	type PostVoteRequest struct {
		Id     int64  `json:"id"`
		Action string `json:"action"` //"like", "dislike", "neutral"
	}

	vote := &PostVoteRequest{}
	err := jsonRequestExtractor(request, vote)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "vote post: "+err.Error())
		return
	}

	err, _ = db.React(int64(vote.Id), activeUser.ID, "post", vote.Action)
	if err != nil && !errors.Is(err, custom_errs.ErrNotificationFailed) {
		if !errors.Is(err, custom_errs.ErrUserNotConnected) {
			fmt.Println("Database update failed:", err)
			jsonProblemResponder(writer, http.StatusOK, "", "VotePost: Failed to update post likes/dislikes. "+err.Error())
			return
		}
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("VotePost: Failed to send final response!")
	}
}
