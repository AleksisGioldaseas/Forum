package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"forum/server/core/config"
	"html/template"
	"io"
	"net/http"
	"path"
	"strings"
)

type GenericResponse struct {
	IsLoggedIn        bool
	ActiveUsername    string
	ActiveProfilePic  *string
	NotificationAlert bool
	NotificationCount int
	ActiveUserRole    int
}

type JsonResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ConsoleLog string `json:"console_log"`
	Data       any    `json:"data"`
}

// responds with a standardized fail json package and adds a messages to it, one for user and one for console output
func jsonProblemResponder(writer http.ResponseWriter, statusCode int, userMessage, consoleMessage string) error {
	if userMessage == "" {
		userMessage = "Something went wrong"
	}
	fmt.Println("json problem found:", userMessage, consoleMessage)
	response := JsonResponse{
		Success:    false,
		Message:    userMessage,
		ConsoleLog: consoleMessage,
	}
	writer.WriteHeader(statusCode)
	return json.NewEncoder(writer).Encode(response)

}

// responds with a standardized success json package and adds the data to it, you can add an optional message if you want a green popup to appear
func jsonOkResponder(writer http.ResponseWriter, data any, optionalMessages ...string) error {
	response := JsonResponse{
		Success: true,
		Data:    data,
		Message: strings.Join(optionalMessages, ","),
	}
	writer.WriteHeader(http.StatusOK)
	fmt.Println("sending an ok!")
	return json.NewEncoder(writer).Encode(response)
}

// gets the data struct from the request
func jsonRequestExtractor[T any](request *http.Request, requestFormat *T) error {
	if request.Header.Get("Content-Type") != "application/json" {
		return errors.New("jsonRequestExtractor: Incorrect content type header")
	}

	// 2. Read and CLONE the body upfront
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	request.Body.Close() // Explicitly close original

	// 3. Restore the body for potential reuse
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 4. Decode from the CLONED bytes (not the stream)
	if err := json.Unmarshal(bodyBytes, requestFormat); err != nil {
		return errors.New("Unable to decode json. " + err.Error())
	}

	return nil
}

type ErrorDataT struct {
	GenericResponse
	//all pages
	IsLoggedIn       bool
	ActiveUsername   string
	ActiveProfilePic *string

	Message string
}

func showErrorPage(writer http.ResponseWriter, activeUser config.User, activeUserIsLoggedIn bool, status int, message string) error {
	writer.WriteHeader(status)

	// Build the correct path to the template
	templatePath := path.Join("web", "templates", "error.html")
	modulesTemplatePath := path.Join("web", "static", "modules.html")

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath, modulesTemplatePath)
	if err != nil {
		fmt.Println("error parsing error page", err)
		return fmt.Errorf("error parsing template: %v", err)
	}

	data := ErrorDataT{
		IsLoggedIn:       activeUserIsLoggedIn,
		ActiveUsername:   activeUser.UserName,
		ActiveProfilePic: activeUser.ProfilePic,

		Message: message,
	}

	if err := tmpl.ExecuteTemplate(writer, "error.html", data); err != nil {
		fmt.Println("error executing >>> ERROR <<< page: ", err)
		return fmt.Errorf("template execution error: %v", err)
	}

	return nil
}
