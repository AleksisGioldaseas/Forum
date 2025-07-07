package sse

import (
	"fmt"
	"forum/common/custom_errs"
	"sync"
	"time"
)

const (
	CLOSE = iota
	NOTIFY
)

var userConnections = make(map[int64]chan int)
var mutex = sync.RWMutex{}

func NewConnection(userId int64, newChannel chan int) {
	mutex.Lock()
	defer mutex.Unlock()

	oldChannel, ok := userConnections[userId]

	if ok {
		select {
		case oldChannel <- CLOSE:
		case <-time.After(1 * time.Second):
			fmt.Println("-#-     ERROR:FAILED TO CLOSE OLD CONNECTION")
		}
	}

	userConnections[userId] = newChannel

}

func EndConnection(userId int64, channel chan int) {
	mutex.Lock()
	defer mutex.Unlock()

	c, ok := userConnections[userId]
	if ok && c == channel {
		delete(userConnections, userId)
	}
}

func SendSSENotification(userId int64) error {
	if userId == 0 {
		return nil
	}
	mutex.RLock()
	channel, ok := userConnections[userId]
	mutex.RUnlock()

	if !ok {
		return custom_errs.ErrUserNotConnected
	}

	select {
	case channel <- NOTIFY:
		return nil

	case <-time.After(1 * time.Second):
		return fmt.Errorf("notification timed out (channel blocked) %w", custom_errs.ErrGeneralSse)
	}
}
