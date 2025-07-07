package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func DeleteCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("deleteComment Handler called!")

	type deleteCommentRequest struct {
		CommentId int64 `json:"comment_id"`
	}

	deleteCommentReq := &deleteCommentRequest{}
	err := jsonRequestExtractor(request, deleteCommentReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", err.Error())
		return
	}

	err = db.DeleteCom(activeUser.ID, deleteCommentReq.CommentId)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Database error: "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("DeleteCommentHandler: Failed to send final response!")
	}

}
