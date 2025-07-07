package handlers

import (
	"encoding/json"
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"html/template"
	"log"
	"net/http"
	"path"
	"strconv"
)

func SuperReportPageFormHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("CreatePost page handler called!")

	activeUserIsLoggedIn := *activeUser != config.User{}

	urlData := request.URL.Query()
	commentIdStr := urlData.Get("comment_id")
	postIdStr := urlData.Get("post_id")
	userIdStr := urlData.Get("user_id")

	var commentId int
	var postId int
	var userId int

	var err error

	switch {
	case commentIdStr != "":
		commentId, err = strconv.Atoi(commentIdStr)

	case postIdStr != "":
		postId, err = strconv.Atoi(postIdStr)

	case userIdStr != "":
		userId, err = strconv.Atoi(userIdStr)
	}

	if err != nil {
		fmt.Println("Bad url variable id given for super report target: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	var Comment *config.Comment
	var Post *config.Post
	var User *config.User

	tx, _ := db.ExportTx()
	defer tx.Rollback()

	switch {
	case commentId != 0:
		Comment, err = db.GetCommentById(tx, activeUser.Role, int64(commentId))
	case postId != 0:
		Post, err = db.GetPostById(tx, activeUser.ID, int64(postId), true)
	case userId != 0:
		User, err = db.GetUserById(tx, int64(userId))
	}

	if err != nil {
		fmt.Println("Can't find report target: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - Report Target Not Found")
		return
	}

	hasUnseen, notificationCount, err := db.CountNotifications(nil, tx, int(activeUser.ID))
	if err != nil {
		fmt.Println("Error unable to count notifiations: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Print("SuperReportPageFormHandler: Transaction Failure", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	postJson, err := json.Marshal(Post)
	if err != nil {
		fmt.Println("unable to marshan target postJson: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}
	commentJson, err := json.Marshal(Comment)
	if err != nil {
		fmt.Println("unable to marshan target commentJson: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}
	userJson, err := json.Marshal(User)
	if err != nil {
		fmt.Println("unable to marshan target userJson: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	type SuperReportFormDataT struct {
		//all pages
		GenericResponse

		//specific
		Comment *config.Comment
		Post    *config.Post
		User    *config.User

		JsonComment string
		JsonPost    string
		JsonUser    string
	}

	data := SuperReportFormDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},

		Comment:     Comment,
		Post:        Post,
		User:        User,
		JsonComment: string(commentJson),
		JsonPost:    string(postJson),
		JsonUser:    string(userJson),
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexSuperReportForm.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	if err := tmpl.ExecuteTemplate(writer, "indexSuperReportForm.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}

}
