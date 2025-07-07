package handlers

//handles login requests from the front

import (
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/utils"

	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoginHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Login Handler called!")

	// Timer to prevent timing attacks
	startTime := time.Now()

	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	loginReq := &LoginRequest{}
	err := jsonRequestExtractor(request, loginReq)
	if err != nil {
		jsonProblemResponder(writer, http.StatusBadRequest, "", "Login handler: "+err.Error())
		return
	}

	// Call login
	user, err := login(db, loginReq.Username, loginReq.Password)
	if err != nil {
		//delay here
		waitUntilTimeHasPassedSince(startTime)
		switch {
		case errors.Is(err, custom_errs.ErrNameNotFound):
			jsonProblemResponder(writer, http.StatusNotFound, "Wrong credentials", "Wrong credentials")
		case errors.Is(err, custom_errs.ErrInvalidPassword):
			jsonProblemResponder(writer, http.StatusUnauthorized, "Wrong credentials", "Wrong credentials")
		case errors.Is(err, custom_errs.ErrDBConnetionIsLost):
			jsonProblemResponder(writer, http.StatusInternalServerError, "", "login db error: "+err.Error())
		default:
			fmt.Println(err)
			jsonProblemResponder(writer, http.StatusInternalServerError, "", "login unexpected err: "+err.Error())
		}
		return
	}

	// Check if there is an open session
	hasSession, err := db.HasSession(request)
	if err != nil {
		waitUntilTimeHasPassedSince(startTime)
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Login: Failed to check active sessions: "+err.Error())
		return
	}
	if hasSession {
		waitUntilTimeHasPassedSince(startTime)
		jsonProblemResponder(writer, http.StatusConflict, "You're already logged in", "user already logged in trying to login again")
		return
	}

	// Generate UUID session token
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	err = db.StoreSession(sessionToken, user.ID, expiresAt)
	if err != nil {
		waitUntilTimeHasPassedSince(startTime)
		jsonProblemResponder(writer, http.StatusInternalServerError, "", "Login: Failed to store session. "+err.Error())
		return
	}

	// Bake cookie
	http.SetCookie(writer, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(Configuration.CookieExpirationHours.ToDuration()),
		HttpOnly: true, // ATTENTION: This prevents client side JS access
		Secure:   true,
		Path:     "/",                     // This makes the cookie available to the entire site
		SameSite: http.SameSiteStrictMode, // This prevents CSRF apparently
	})

	waitUntilTimeHasPassedSince(startTime)
	err = jsonOkResponder(writer, nil)
	if err != nil {
		fmt.Println("Login: Failed to send final response!")
	}
}

// waits until a specific amount of time has passed since the argument, to stop timing attacks on login
func waitUntilTimeHasPassedSince(start time.Time) {
	elapsedTime := time.Since(start)
	minTime := 500 * time.Millisecond
	if elapsedTime < minTime {
		time.Sleep(minTime - elapsedTime)
	}
}

// executes the login procedure
func login(db *database.DataBase, username, password string) (*config.User, error) {
	// Sanitise inputs
	username = database.Sanitize(username)

	user, err := db.GetAuthUserByUserName(username)
	if err != nil {
		return nil, fmt.Errorf("login: problem fidning auth user: %w", err)
	}

	hashedPass := utils.HashPass(password, user.Salt, Configuration.XorKey)

	if hashedPass != user.PasswordHash {
		return nil, custom_errs.ErrInvalidPassword
	}

	return user, nil
}
