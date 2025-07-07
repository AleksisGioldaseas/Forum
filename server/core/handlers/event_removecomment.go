package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func RemoveCommentHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("removeComment Handler called!")

	type removeCommentRequest struct {
		CommentId int64  `json:"comment_id"`
		Reason    string `json:"reason"`
	}

	removeCommentReq := &removeCommentRequest{}
	err := jsonRequestExtractor(request, removeCommentReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "remove comment: "+err.Error())
		return
	}

	err = db.ModAssessment("comment", removeCommentReq.CommentId, activeUser.ID, removeCommentReq.Reason, activeUser.UserName, 1)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "failed to remove comment. "+err.Error(), "console")
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("RemoveCommentHandler: Failed to send final response!")
	}

}
