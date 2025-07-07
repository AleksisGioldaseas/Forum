package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Google login handler called")
	state, err := utils.GenerateStateCookie()
	if err != nil {
		jsonProblemResponder(w, http.StatusInternalServerError, "", "google login: failed to generate state")
		return
	}
	utils.SetStateCookie(w, state)

	url := config.GoogleOAuth.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Google callback handler called")
	state := r.FormValue("state")
	cookieState, err := utils.GetStateCookie(r)
	if err != nil || state != cookieState {
		fmt.Println(custom_errs.ErrOAuthStateMismatch.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusBadRequest, "ERROR 400: Bad Request")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	code := r.FormValue("code")
	accessToken, err := config.GoogleOAuth.ExchangeCodeForToken(code)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthCodeExchange.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	request, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthInfoRequest.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	re, err := client.Do(request)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthUserInfoFetch.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusServiceUnavailable, "ERROR 500: Service Unavailable")
		return
	}
	defer re.Body.Close()

	if re.StatusCode != http.StatusOK {
		fmt.Printf("Google API returned status: %s\n", re.Status)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusServiceUnavailable, "ERROR 500: Service Unavailable")
		return
	}

	var userInfo struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
		Verified bool   `json:"verified_email"`
		Sub      string `json:"id"`
	}

	if err := json.NewDecoder(re.Body).Decode(&userInfo); err != nil {
		fmt.Printf("failed to decode user info: %v\n", err)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	if !userInfo.Verified {
		fmt.Println(custom_errs.ErrOAuthEmailNotVerified.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusForbidden, "ERROR 403: Forbidden\nEmail not verified")
		return
	}

	user, err := db.GetUserBySub(userInfo.Sub)
	if err != nil {
		if errors.Is(err, custom_errs.ErrNameNotFound) {
			username := utils.GenerateUsername(userInfo.Name)
			if username == "" {
				username = utils.GenerateUsername(userInfo.Email)
			}

			oauthPass, err := utils.GenerateOAuthPassword()
			if err != nil {
				fmt.Println("failed to generate OAuth password")
				showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
				return
			}
			userSalt, err := utils.Salt()
			if err != nil {
				fmt.Println("failed to generate salt")
				showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
				return
			}
			hashedPass := utils.HashPass(oauthPass, userSalt, Configuration.XorKey)

			picLink, err := saveImage(userInfo.Picture)
			if err != nil {
				fmt.Printf("failed to add external image: %v", err)
				picLink = "default_pfp.jpg"
			}

			err = db.AddImage(nil, picLink)
			if err != nil {
				fmt.Printf("failed to add external image: %v", err)
				picLink = "default_pfp.jpg"
			}

			newUser := &config.User{
				UserName:     username,
				Email:        userInfo.Email,
				PasswordHash: hashedPass,
				ProfilePic:   &picLink,
				Bio:          nil,
				Role:         1,
				OAuthSub:     userInfo.Sub,
				Salt:         userSalt,
			}

			id, err := db.AddOAuthUser(newUser)

			if err != nil {
				fmt.Println(err)
				if errors.Is(err, custom_errs.ErrEmailNotUnique) {
					fmt.Println(custom_errs.ErrEmailNotUnique.Error())
					showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusBadRequest,
						"ERROR 400: Bad Request\nThe email associated with this account already exists")
				} else {
					fmt.Println(custom_errs.ErrUserCreationFailed.Error())
					showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError,
						"ERROR 500: Internal Error")
				}
				return
			}

			user = &config.User{
				ID:         id,
				UserName:   username,
				Email:      userInfo.Email,
				ProfilePic: &userInfo.Picture,
			}
		} else {
			fmt.Printf("failed to get user: %v\n", err)
			showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
			return
		}
	}

	hasSession, err := db.HasSession(r)
	if err != nil {
		fmt.Printf("Failed to check sessions: %v\n", err)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}
	if hasSession {
		fmt.Println(custom_errs.ErrSessionAlreadyExists.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusConflict, "ERROR 409: Conflict\nUser already has an active session")
		return
	}

	// Create the session
	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	err = db.StoreSession(sessionToken, user.ID, expiresAt)
	if err != nil {
		fmt.Println(custom_errs.ErrSessionCreation.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	// Bake cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func GithubLoginHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Github handler called")
	state, err := utils.GenerateStateCookie()
	if err != nil {
		jsonProblemResponder(w, http.StatusInternalServerError, "", "github login: failed to generate state")
		return
	}
	utils.SetStateCookie(w, state)

	url := config.GithubOAuth.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GithubCallbackHandler(w http.ResponseWriter, r *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("Github callback handler called")
	state := r.FormValue("state")
	cookieState, err := utils.GetStateCookie(r)
	if err != nil || state != cookieState {
		fmt.Println(custom_errs.ErrOAuthStateMismatch.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusBadRequest, "ERROR 400: Bad Request")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	code := r.FormValue("code")
	accessToken, err := config.GithubOAuth.ExchangeCodeForToken(code)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthCodeExchange.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	userReq, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthInfoRequest.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}
	userReq.Header.Set("Authorization", "Bearer "+accessToken)
	userReq.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}

	// First request to get basic user info
	userRe, err := client.Do(userReq)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthUserInfoFetch.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusServiceUnavailable, "ERROR 500: Service Unavailable")
		return
	}
	defer userRe.Body.Close()

	if userRe.StatusCode != http.StatusOK {
		fmt.Printf("Github API returned status: %s\n", userRe.Status)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusServiceUnavailable, "ERROR 500: Service Unavailable")
		return
	}

	var userInfo struct {
		Login   string `json:"login"`
		Name    string `json:"name"`
		Picture string `json:"avatar_url"`
		ID      int    `json:"id"`
	}

	if err := json.NewDecoder(userRe.Body).Decode(&userInfo); err != nil {
		fmt.Printf("failed to decode user info: %v\n", err)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	// Second request to ensure we got primary email
	emailReq, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		fmt.Printf("failed to create email request: %v\n", err)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}
	emailReq.Header.Set("Authorization", "Bearer "+accessToken)
	emailReq.Header.Set("Accept", "application/json")

	emailRe, err := client.Do(emailReq)
	if err != nil {
		fmt.Println(custom_errs.ErrOAuthUserInfoFetch.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusServiceUnavailable, "ERROR 500: Service Unavailable")
		return
	}
	defer emailRe.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(emailRe.Body).Decode(&emails); err != nil {
		fmt.Println("no primary email found")
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusForbidden, "ERROR 403: Forbidden\nNo primary email associated with this Github account")
		return
	}

	var primaryEmail string
	var verifiedEmail bool
	for _, email := range emails {
		if email.Primary {
			primaryEmail = email.Email
			verifiedEmail = email.Verified
			break
		}
	}

	if primaryEmail == "" {
		fmt.Println("no primary email found")
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusForbidden, "ERROR 403: Forbidden\nNo primary email associated with this Github account")
		return
	}

	if !verifiedEmail {
		fmt.Println(custom_errs.ErrOAuthEmailNotVerified.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusForbidden, "ERROR 403: Forbidden\nPrimary email not verified")
		return
	}

	gitSub := fmt.Sprintf("%d", userInfo.ID)

	user, err := db.GetUserBySub(gitSub)
	if err != nil {
		if errors.Is(err, custom_errs.ErrNameNotFound) {
			username := utils.GenerateUsername(userInfo.Login)
			if username == "" {
				username = utils.GenerateUsername(primaryEmail)
			}

			oauthPass, err := utils.GenerateOAuthPassword()
			if err != nil {
				fmt.Println("failed to generate OAuth password")
				showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
				return
			}
			userSalt, err := utils.Salt()
			if err != nil {
				fmt.Println("failed to generate salt")
				showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
				return
			}
			hashedPass := utils.HashPass(oauthPass, userSalt, Configuration.XorKey)

			picLink, err := saveImage(userInfo.Picture)
			if err != nil {
				fmt.Printf("failed to add external image: %v", err)
				picLink = "default_pfp.jpg"
			}

			err = db.AddImage(nil, picLink)
			if err != nil {
				fmt.Printf("failed to add external image: %v", err)
				picLink = "default_pfp.jpg"
			}

			newUser := &config.User{
				UserName:     username,
				Email:        primaryEmail,
				PasswordHash: hashedPass,
				ProfilePic:   &picLink,
				Bio:          nil,
				Role:         1,
				OAuthSub:     gitSub,
				Salt:         userSalt,
			}

			id, err := db.AddOAuthUser(newUser)

			if err != nil {
				fmt.Println(err)
				if errors.Is(err, custom_errs.ErrEmailNotUnique) {
					fmt.Println(custom_errs.ErrEmailNotUnique.Error())
					showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 400: Bad Request\nThe email associated with this account already exists")
				} else {
					fmt.Println(custom_errs.ErrUserCreationFailed.Error())
					showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
				}
				return
			}

			user = &config.User{
				ID:         id,
				UserName:   username,
				Email:      primaryEmail,
				ProfilePic: &userInfo.Picture,
			}

		} else {
			fmt.Printf("failed to fetch user: %v\n", err)
			showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
			return
		}
	}

	hasSession, err := db.HasSession(r)
	if err != nil {
		fmt.Printf("failed to check sessions: %v\n", err)
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	if hasSession {
		fmt.Println(custom_errs.ErrSessionAlreadyExists.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusConflict, "ERROR 409: Conflict\nUser already has an active session")
		return
	}

	sessionToken := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	err = db.StoreSession(sessionToken, user.ID, expiresAt)
	if err != nil {
		fmt.Println(custom_errs.ErrSessionCreation.Error())
		showErrorPage(w, *activeUser, activeUser.ID != 0, http.StatusInternalServerError, "ERROR 500: Internal Error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func saveImage(link string) (string, error) {
	fmt.Println("LINK: ", link)
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Referer", "https://accounts.google.com/") // Optional but can help

	// Perform the request
	response, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("profile pic get failed: status: %d", response.StatusCode)
	}

	cfg := Configuration.Images

	filename := uuid.New().String() + ".png"
	var pathAndFile []string
	pathAndFile = append(pathAndFile, cfg.PathPrefix...)
	pathAndFile = append(pathAndFile, filename)
	path := filepath.Join(pathAndFile...)

	dst, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("unable to save file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, response.Body); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filename, nil
}
