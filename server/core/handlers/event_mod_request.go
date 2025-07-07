package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
	"strings"
)

func ModRequestHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("mod request handler called")

	err := db.ModeratorRequest(activeUser.ID, activeUser.UserName)
	if err != nil {
		if strings.Contains(err.Error(), "already submitted") {
			jsonProblemResponder(writer, http.StatusInternalServerError, "You've already requested, try again much later", "mod request: "+err.Error())
		} else {
			jsonProblemResponder(writer, http.StatusInternalServerError, "", "mod request: "+err.Error())
		}

		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data, "Your mod request has been received")
	if err != nil {
		fmt.Println("modRequestHandler: Failed to send final response!")
	}
}
