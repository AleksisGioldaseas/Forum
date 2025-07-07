package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func DeletePostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("deletePost Handler called!")

	type deletePostRequest struct {
		PostId int64 `json:"post_id"`
	}

	deletePostReq := &deletePostRequest{}
	err := jsonRequestExtractor(request, deletePostReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", err.Error())
		return
	}

	err = db.DeletePost(activeUser.ID, deletePostReq.PostId)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Database error: "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("DeletePostHandler: Failed to send final response!")
	}

}
