package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"forum/common/custom_errs"
	"forum/persistence/database"
	"forum/server/core/config"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type writer = http.ResponseWriter
type request = *http.Request

// Security function that authorizes the use of handlers and implements a rate limit
func AuthMiddleware(
	requiredRole int,
	writer writer,
	request request,
	db *database.DataBase,
	nextHandler func(writer, request, *database.DataBase, *config.User),
	rateLimitMaxRequests int,
	ratelimitSecondsInterval float64,
	universalRateLimitRequests int,
	universalRatelimitSecondsInterval float64,
	allowBanned bool) {

	remoteAddr, _, _ := net.SplitHostPort(request.RemoteAddr)

	rateLimitTag := fmt.Sprint(remoteAddr, request.URL.Path)

	//General throttling
	if BlockRequest(int64(universalRateLimitRequests), int64(universalRatelimitSecondsInterval*1000), remoteAddr) {
		fmt.Println("blocking general: ", remoteAddr)
		return
	}

	if BlockRequest(int64(rateLimitMaxRequests), int64(ratelimitSecondsInterval*1000), rateLimitTag) {
		fmt.Println("blocking specific: ", rateLimitTag)
		return
	}

	activeUser, err := GetUserDataFromCookie(request, db)
	if err != nil {
		activeUser = &config.User{Role: GUEST}
	}

	//Error response handling
	isJson := request.Header.Get("Content-Type") == "application/json"

	if activeUser.Role < requiredRole {
		if isJson {
			jsonProblemResponder(writer, http.StatusForbidden, "Naughty naughty! You're not allowed to do that!", "middleware: no autherization for that action")
		} else {
			showErrorPage(writer, *activeUser, activeUser.Role > GUEST, http.StatusForbidden, "Access Denied")
		}
		return
	}

	if !allowBanned && (activeUser.Banned == 1) {
		if isJson {
			jsonProblemResponder(writer, http.StatusForbidden, "You can't do that 'cause you banned", "middleware: no autherization for that action, reason: ban")
		} else {
			showErrorPage(writer, *activeUser, activeUser.Role > GUEST, http.StatusForbidden, "Access Denied Due to Banned Status")
		}
		return
	}

	nextHandler(writer, request, db, activeUser)

	// should stop panics happening in handlers, so that errors don't take down the server, hopefully avoiding rare bugs from taking down the server
	defer func() {
		r := recover()
		if r != nil {
			if isJson {
				jsonProblemResponder(writer, http.StatusInternalServerError, "", "something very wrong that we won't disclose!")
			} else {
				showErrorPage(writer, *activeUser, activeUser.Role > GUEST, http.StatusInternalServerError, "Sorry, something went really wrong.")
			}

			for range 10 {
				fmt.Fprintln(os.Stderr, "UNEXPECTED PANIC IN AUTH-MIDDLEWARE:", r)
			}
		}
	}()
}

type requestEntry struct {
	count              atomic.Int64
	timeSinceLastReset atomic.Int64
}

var keyAttempts = initSyncMap()

func initSyncMap() *sync.Map {
	var m sync.Map
	return &m
}

func syncMapCleaner() {
	for {
		time.Sleep(time.Hour)
		time.Sleep(time.Minute * time.Duration(rand.IntN(20)))
		keyAttempts.Clear()
	}
}

func BlockRequest(limit int64, intervalMilli int64, key string) bool {
	entryAny, ok := keyAttempts.Load(key)

	if !ok {
		//we do a second check to hopefully catch simultaneous loads from multiple goroutines
		newEntry := requestEntry{}
		newEntry.count.Store(1)
		newEntry.timeSinceLastReset.Store(time.Now().UnixMilli())
		entryAny, ok = keyAttempts.LoadOrStore(key, &newEntry)

		if !ok {
			return false
		}
	}

	entry := entryAny.(*requestEntry)

	//how many times we're overdue for a reset
	count := entry.count.Load()
	timesOver := (time.Now().UnixMilli() - entry.timeSinceLastReset.Load()) / intervalMilli

	if timesOver > 0 {
		//reduce the counter based on how many resets should have happened since the last check
		entry.count.Store(max(1, count-(timesOver*limit)))
		entry.timeSinceLastReset.Store(time.Now().UnixMilli())
	} else {
		entry.count.Add(1)
	}

	//check if we're past the limit
	if entry.count.Load() > limit {
		fmt.Println(entry.count.Load(), " more than ", limit)
		return true
	}

	return false
}

// Gets the user's data from the cookie's session token
func GetUserDataFromCookie(request *http.Request, db *database.DataBase) (*config.User, error) {
	errMsg := "get user data from cookie: %w"
	u := config.NewUser()

	cookie, err := request.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return u, custom_errs.ErrSessionNotFound
		}
		return u, fmt.Errorf("error reading cookie: %w", err)
	}

	u, err = db.GetUserByCookie(cookie.Value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, custom_errs.ErrSessionNotFound
		}
		return u, fmt.Errorf(errMsg, err)
	}
	return u, nil
}
