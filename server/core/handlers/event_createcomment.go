package handlers

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func CreateCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Create comment handler called!")

	type CommentRequest struct {
		PostID int64  `json:"post_id"`
		Body   string `json:"body"`
	}

	commentReq := &CommentRequest{}
	err := jsonRequestExtractor(request, commentReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", err.Error())
		return
	}

	sanitizedBody := string(database.BasicHTMLSanitize(commentReq.Body))

	comment := config.Comment{
		UserID: activeUser.ID, //Setting user id ourselves because we can't trust the user
		PostID: commentReq.PostID,
		Body:   sanitizedBody,
	}

	err = db.ValidateCommentBody(comment.Body)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Comment length is not ok", "CreateComment: Failed to create comment. "+err.Error())
		return
	}

	_, err = db.AddComment(&comment)
	if err != nil && !errors.Is(err, custom_errs.ErrNotificationFailed) {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "CreateComment: Failed to create comment. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("CreateComment: Failed to send final response!")
	}
}
