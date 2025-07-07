package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func EditPostHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("EditPost Handler called!")

	type editPostRequest struct {
		PostId int64  `json:"post_id"`
		Body   string `json:"body"`
	}

	editPostReq := &editPostRequest{}
	err := jsonRequestExtractor(request, editPostReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", err.Error())
		return
	}

	sanitizedBody := string(database.BasicHTMLSanitize(editPostReq.Body))

	err = db.ValidatePostBody(sanitizedBody)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, err.Error(), "editpost: "+err.Error())
		return
	}

	err = db.EditPost(editPostReq.PostId, sanitizedBody, activeUser.ID)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "failed to edit post. "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("EditPostHandler: Failed to send final response!")
	}

}
