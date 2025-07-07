package handlers

//handler for a post

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
	"strings"
)

func PostPageHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Post page handler called!@")

	activeUserIsLoggedIn := *activeUser != config.User{}

	temp := strings.ToLower(strings.TrimPrefix(request.URL.Path, "/post/"))
	temp = strings.TrimSuffix(temp, "/")
	parts := strings.Split(temp, "/")
	postIdstr := parts[0]

	postId, err := strconv.Atoi(postIdstr)
	if err != nil {
		fmt.Println("non number post id given by user")
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - Page Not Found")
		return
	}

	// not caching the err because if err the tx will be nil and all funcs roll back to db.conn
	tx, _ := db.ExportTx()
	defer tx.Rollback()

	post, err := db.GetPostById(tx, activeUser.ID, int64(postId), activeUser.Role > 1) // added user id here to and changed postId to int64 to run db func
	if err != nil || post.IsSuperReport {

		fmt.Println("unable to find post using id: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusNotFound, "404 - Page Not Found")
		return
	}

	mainPostJson, err := json.Marshal(post)
	if err != nil {
		fmt.Println("unable to marshal main post: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusBadRequest, "400 - Bad Request")
		return
	}

	commentsStruct := struct { //commentStruct
		Comments []*config.Comment `json:"comments"`
	}{}
	var errs []error
	commentsStruct.Comments, errs = db.GetComments(tx, int(post.ID), 5, 0, activeUser.ID, activeUser.Role, "old", false, false)
	if len(errs) > 0 {
		fmt.Println("errors fetching comments by post id: ", errs)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	commentsJson, err := json.Marshal(commentsStruct)
	if err != nil {
		fmt.Println("unable to marshal main post: ", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
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
		log.Print("post page handler: Transaction Failure", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	type PostDataT struct {
		//all pages
		GenericResponse

		//specific
		PostId       int
		JsonMainPost string
		MainPost     *config.Post
		JsonComments string
		Comments     []*config.Comment
	}

	data := PostDataT{
		GenericResponse: GenericResponse{
			IsLoggedIn:        activeUserIsLoggedIn,
			ActiveUsername:    activeUser.UserName,
			ActiveProfilePic:  activeUser.ProfilePic,
			NotificationAlert: hasUnseen,
			NotificationCount: notificationCount,
			ActiveUserRole:    activeUser.Role,
		},

		PostId:       int(post.ID),
		JsonMainPost: string(mainPostJson),
		MainPost:     post,
		JsonComments: string(commentsJson),
		Comments:     commentsStruct.Comments,
	}

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "indexPost.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		// SendErrorPage(writer, 500, "500 - Internal Server Error")
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	if err := tmpl.ExecuteTemplate(writer, "indexPost.html", data); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, *activeUser, activeUserIsLoggedIn, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
