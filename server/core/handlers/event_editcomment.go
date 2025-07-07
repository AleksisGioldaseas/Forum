package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func EditCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("EditComment Handler called!")

	type editCommentRequest struct {
		CommentId int64  `json:"comment_id"`
		Body      string `json:"body"`
	}

	editCommentReq := &editCommentRequest{}
	err := jsonRequestExtractor(request, editCommentReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Something went wrong", "edit comment: "+err.Error())
		return
	}

	sanitizedBody := string(database.BasicHTMLSanitize(editCommentReq.Body))

	err = db.ValidateCommentBody(sanitizedBody)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Comment length is not ok", "EditComment: Failed to create comment. "+err.Error())
		return
	}

	err = db.EditComment(editCommentReq.CommentId, sanitizedBody, activeUser.ID)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "failed to edit comment. "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("EditCommentHandler: Failed to send final response!")
	}

}
