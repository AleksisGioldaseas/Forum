package handlers

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/server/core/sse"
	"net/http"
)

// this accepts an sse connection from webpage
func NotificationSSEHandler(writer http.ResponseWriter, request *http.Request, db *database.DataBase, activeUser *config.User) {
	fmt.Println("NotificationSSEHandler called!")
	defer fmt.Println("Notification handler closing!")
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	notificationChannel := make(chan int)

	sse.NewConnection(activeUser.ID, notificationChannel)

	//===== SEND NOTIF ====
	hasUnseen, count, err := db.CountNotifications(nil, nil, int(activeUser.ID))
	if err != nil {
		return
	}
	if hasUnseen {
		fmt.Fprintf(writer, "data: %s\n\n", fmt.Sprint(count))
		flusher.Flush()
	}
	//=====================

	for {
		select {
		case <-request.Context().Done(): //webpage closed its side of the connection

			sse.EndConnection(activeUser.ID, notificationChannel)
			fmt.Println("context done from request")
			return

		case action, ok := <-notificationChannel:
			if !ok { // Channel closed by sender
				panic(2)
				// return
			}
			if action == sse.NOTIFY {
				//===== SEND NOTIF ====
				ctx := request.Context()
				hasUnseen, count, err := db.CountNotifications(&ctx, nil, int(activeUser.ID))
				if err != nil {
					fmt.Println("count notif err")
					return
				}
				if hasUnseen {
					fmt.Fprintf(writer, "data: %s\n\n", fmt.Sprint(count))
					flusher.Flush()
				}
				//=====================

				continue
			}
			sse.EndConnection(activeUser.ID, notificationChannel)
			close(notificationChannel)
			fmt.Println("exit received")
			return
		}
	}

}
