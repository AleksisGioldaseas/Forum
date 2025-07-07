package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func BanUserHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("BanUser Handler called!")

	type BanUserRequest struct {
		Username string `json:"username"`
		Days     int64  `json:"days"`
		Reason   string `json:"reason"`
	}

	BanUserReq := &BanUserRequest{}
	err := jsonRequestExtractor(request, BanUserReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "ban user: "+err.Error())
		return
	}

	sanitizedUsername := database.Sanitize(BanUserReq.Username)
	sanitizedReason := database.Sanitize(BanUserReq.Reason)

	if BanUserReq.Days < 1 || BanUserReq.Days > 999999 {
		jsonProblemResponder(writer, http.StatusBadRequest, "Invalid ban duration (1-999999 days)", "ban user: invalid duration")
		return
	}

	err = db.BanUser(sanitizedUsername, BanUserReq.Days, sanitizedReason, activeUser)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to ban user", "failed to ban user: "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("BanUserHandler: Failed to send final response!")
	}

}

func UnBanUserHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {

	fmt.Println("UnBanUser Handler called!")

	type UnBanUserRequest struct {
		Username string `json:"username"`
	}

	UnBanUserReq := &UnBanUserRequest{}
	err := jsonRequestExtractor(request, UnBanUserReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "Bad request", "unban:"+err.Error())
		return
	}

	sanitizedUsername := database.Sanitize(UnBanUserReq.Username)

	err = db.UnBanUser(sanitizedUsername)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "Failed to unban", "failed to unban user. "+err.Error())
		return
	}

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("BanUserHandler: Failed to send final response!")
	}

}
