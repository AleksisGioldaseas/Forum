package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func PromoteUserHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Promote user handler called!")

	type promoteUserRequest struct {
		Username string `json:"username"`
	}

	promoteUserReq := &promoteUserRequest{}
	err := jsonRequestExtractor(request, promoteUserReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "promote user:"+err.Error())
		return
	}

	err = db.UpdateRole(promoteUserReq.Username, activeUser.ID, MOD)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to update role", "promote user: "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("PromoteUserHandler: Failed to send final response!")
	}

}
