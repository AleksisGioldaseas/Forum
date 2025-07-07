package handlers

//handles login requests from the front

import (
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"net/http"
	"time"
)

func LogoutHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Logout Handler called!")

	cookie, err := request.Cookie("session_token")
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "You're not logged in", custom_errs.ErrSessionNotFound.Error())
		return
	}

	sessionToken := cookie.Value

	// Delete session from the db
	err = db.DeleteSession(sessionToken)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "logout: "+err.Error())
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("Postlist: Failed to send final response!")
	}
}

func LogoutAllHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Logout All handler called!")

	cookie, err := request.Cookie("session_token")
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "You're not logged in", "logout all: "+custom_errs.ErrSessionNotFound.Error())
		return
	}

	currentSession := cookie.Value

	otherSessions, err := db.GetOtherSessions(activeUser.ID, currentSession)
	if err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "logout all: "+err.Error())
		return
	}

	if len(otherSessions) == 0 {
		jsonOkResponder(writer, nil)
		return
	}

	if err := db.DeleteSessions(otherSessions); err != nil {
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "logout all: "+err.Error())
		return
	}

	jsonOkResponder(writer, map[string]any{
		"message":       "logged out from all other devices",
		"sessionsEnded": len(otherSessions),
	})
}
