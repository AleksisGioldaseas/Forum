package handlers

//handler for home page

import (
	"encoding/json"
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"html/template"
	"log"
	"net/http"
	"path"
)

func NotificationsPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {

	fmt.Println("NotificationsPage handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexNotifications.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	tx, _ := db.ExportTx()
	defer tx.Rollback()

	hasUnseen, notifications, _, unSeenNotificationCount, err := db.GetNotifications(tx, activeUser.ID, 15, 0)
	if err != nil {
		fmt.Println("Error getting notifications: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	notificationsJson, err := json.Marshal(notifications)
	if err != nil {
		fmt.Println("unable to marshal notifications: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	if hasUnseen {
		err = db.AllNotificationsSeen(tx, activeUser.ID)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Print("notifications page handler: Transaction Failure", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	type NotificationPageDataT struct {
		//all pages
		GenericResponse

		NotificationsJson string
	}

	data := NotificationPageDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: unSeenNotificationCount,
			ActiveUserRole:    activeUser.Role,
		},
		NotificationsJson: string(notificationsJson),
	}

	// fmt.Println(data.NotificationAlert)

	if err := tmpl.ExecuteTemplate(writer, "indexNotifications.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
