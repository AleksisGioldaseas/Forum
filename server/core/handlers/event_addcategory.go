package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func AddCategoryHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Add category handler called!")

	type CategoryRequest struct {
		Name string `json:"name"`
	}

	categoryReq := &CategoryRequest{}
	err := jsonRequestExtractor(request, categoryReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", err.Error())
		return
	}

	_, err = db.AddCategory(categoryReq.Name)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to add category", "Add category: Failed to add category. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("Add category: Failed to send final response!")
	}
}
