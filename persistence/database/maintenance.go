package database

import (
	"fmt"
	"forum/common/custom_errs"
	"forum/utils"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Use this channel to force WAL to Truncate. Usage:
//
//	 database.Wg.Add(1)
//	`ManualTruncate <- struct {}{}`
var ManualTruncate = make(chan struct{})
var Wg = &sync.WaitGroup{}

// Call this func in a go routine with the optimal refresh time duration. Then you can send a signal to ManualTruncate chan at any time to force truncate
// Usage:
//
//	`ManualTruncate <- struct {}{}`
func (db *DataBase) StartTruncateRoutine(interval utils.Duration) {
	var checkPointInProgress int32
	ticker := time.NewTicker(interval.ToDuration())
	Wg.Add(1)
	go func() {

		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				Wg.Add(1)
				shouldCheckpoint, err := db.CheckWALPressure()
				if err == nil && shouldCheckpoint {
					db.truncate("auto", &checkPointInProgress)
				} else {
					Wg.Done()
				}
			case <-ManualTruncate:
				db.truncate("manual", &checkPointInProgress)

			}
		}

	}()
}

func (db *DataBase) truncate(triger string, checkpointInProgress *int32) {
	defer Wg.Done()
	if !atomic.CompareAndSwapInt32(checkpointInProgress, 0, 1) {
		log.Println("DATABASE: Checkpoint already in progress, skipping.")
		return
	}
	var walType string
	switch triger {
	case "auto":
		walType = WAL_ALL
	case "manual":
		walType = TRUNCATE
	}

	_, err := db.conn.Exec(walType)
	if err != nil {
		log.Println(custom_errs.ErrWalCheckpoint, err)
	} else {
		log.Printf("DATABASE: WAL checkpoint completed (%s)", triger)
	}
	atomic.StoreInt32(checkpointInProgress, 0)
}

// CheckWALPressure checks how many WAL frames are pending and advises on checkpoint timing.
func (db *DataBase) CheckWALPressure() (shouldCheckpoint bool, err error) {
	var resultCode, logFrames, checkpointedFrames int

	row := db.conn.QueryRow(WAL_PRESSURE_CHECK)
	if err := row.Scan(&resultCode, &logFrames, &checkpointedFrames); err != nil {
		return false, fmt.Errorf("failed to read WAL status: %w", err)
	}

	pending := logFrames - checkpointedFrames

	fmt.Printf("WAL Status:\n")
	fmt.Printf(" - Log Frames: %d\n", logFrames)
	fmt.Printf(" - Checkpointed Frames: %d\n", checkpointedFrames)
	fmt.Printf(" - Pending Frames: %d\n", pending)
	fmt.Printf(" - Result Code: %d\n", resultCode)

	// Suggest a checkpoint if more than 1000 frames are pending
	shouldCheckpoint = pending > 1000 || resultCode != 0
	fmt.Printf(" Should Check Point: %v\n", shouldCheckpoint)
	return shouldCheckpoint, nil
}

func (db *DataBase) walSetUp(sync, cacheSize string) {
	db.conn.Exec(fmt.Sprintf("PRAGMA synchronous = %s;", sync))
	db.conn.Exec(fmt.Sprintf("PRAGMA cache_size = %s;", cacheSize))
}

func (db *DataBase) cleanupExpiredSessions() {
	log.Println("DATABASE: Cleaning up expired sessions...")
	_, err := db.conn.Exec(CLEAN_UP)
	if err != nil {
		log.Println("DATABASE: Error cleaning up expired sessions:", err)
	} else {
		log.Println("DATABASE: Expired sessions deleted.")
	}
}

func (db *DataBase) sessionCleanupTask(interval utils.Duration) {
	ticker := time.NewTicker(interval.ToDuration())
	Wg.Add(1)

	go func() {
		defer Wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				db.cleanupExpiredSessions()
			case <-db.Ctx.Done():
				log.Println("Shuting down session clean up task")
				return
			}
		}
	}()
}
