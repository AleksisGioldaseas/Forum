package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

// lets the server know which notifications the user has seen
func NotificationsSeenHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("notificationsSeen Handler called!")

	type notificationsSeenRequest struct {
		NotificationIds []int64 `json:"notification_ids"`
	}

	notificationsSeenReq := &notificationsSeenRequest{}
	err := jsonRequestExtractor(request, notificationsSeenReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "see handler: "+err.Error())
		return
	}

	data := struct {
	}{}

	err = jsonOkResponder(writer, data)
	if err != nil {
		fmt.Println("NotificationsSeenHandler: Failed to send final response!")
	}

}
