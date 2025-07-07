package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
)

func GetNotificationListHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("NotificationList Handler called!")

	type notificationListRequest struct {
		Count int `json:"count"`
		Page  int `json:"page"` //page 1, 2, 3 and so on
	}

	notificationsRequest := &notificationListRequest{}
	err := jsonRequestExtractor(request, notificationsRequest)
	if err != nil {
		fmt.Println(err.Error())
		jsonProblemResponder(writer, http.StatusBadRequest, "", "notif list: "+err.Error())
		return
	}

	fmt.Println(notificationsRequest)

	NotificationsStruct := struct { //NotificationStruct
		Notifications []*database.Notification `json:"Notifications"`
	}{}

	_, NotificationsStruct.Notifications, _, _, err = db.GetNotifications(nil, activeUser.ID, notificationsRequest.Count, notificationsRequest.Page)
	if err != nil {
		fmt.Println("errors fetching Notifications:", err)
		showErrorPage(writer, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	err = jsonOkResponder(writer, NotificationsStruct)
	if err != nil {
		fmt.Println("NotificationList: Failed to send final response!")
	}

}
