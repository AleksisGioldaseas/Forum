package handlers

import (
	"forum/persistence/database"
	"forum/server/core/config"
	"html/template"
	"log"
	"net/http"
	"path"
)

func PageTestHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {

	templatePath := path.Join("web", "templates", "test.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		// SendErrorPage(writer, 500, "500 - Internal Server Error")
		log.Printf("Error parsing template: %v", err)
		showErrorPage(writer, config.User{}, false, http.StatusInternalServerError, "500 - Internal Server Error")
		return
	}

	if err := tmpl.ExecuteTemplate(writer, "test.html", nil); err != nil {
		log.Printf("Template execution error: %v", err)
		showErrorPage(writer, config.User{}, false, http.StatusInternalServerError, "500 - Internal Server Error")
	}
}
