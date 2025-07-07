package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func RemoveCategoryHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Remove category handler called!")

	type CategoryRequest struct {
		Name string `json:"name"`
	}

	categoryReq := &CategoryRequest{}
	err := jsonRequestExtractor(request, categoryReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "remove category: "+err.Error())
		return
	}

	err = db.DeleteCategory(categoryReq.Name)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Remove category: Failed to Remove category. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("Remove category: Failed to send final response!")
	}
}
