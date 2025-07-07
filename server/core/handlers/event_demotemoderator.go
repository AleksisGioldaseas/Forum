package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func DemoteModeratorHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Demote mod handler called!")
	type demoteModeratorRequest struct {
		Username string `json:"username"`
	}

	demoteModeratorReq := &demoteModeratorRequest{}
	err := jsonRequestExtractor(request, demoteModeratorReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "demote mod"+err.Error())
		return
	}

	err = db.UpdateRole(demoteModeratorReq.Username, activeUser.ID, USER)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to change role", "demote mod: "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("DemoteModeratorHandler: Failed to send final response!")
	}

}
