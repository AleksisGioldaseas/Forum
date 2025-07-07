package database

import (
	"errors"
	"fmt"
	"forum/server/core/config"
	"log"
	"time"
)

func (db *DataBase) BanUser(targetUsername string, Days int64, Reason string, activeUser *config.User) error {

	errorMsg := "ban user error %w"

	expDate := time.Now().Add(time.Hour * 24 * time.Duration(Days))

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer tx.Rollback()

	var bannedUserId int64
	var bannedUserRole int64

	err = tx.QueryRowContext(
		ctx,
		BAN_USER_BY_NAME,
		expDate,
		Reason,
		activeUser.UserName,
		targetUsername,
	).Scan(&bannedUserId, &bannedUserRole)

	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if bannedUserRole >= int64(activeUser.Role) {
		return fmt.Errorf(errorMsg, errors.New("you don't have permission to ban this user"))
	}

	db.AddNotifGeneric(tx, bannedUserId, 0, "ban", fmt.Sprintf("You've been banned for %d days", Days))

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		go func(username string) {
			_, err := db.GetUserByUserName(nil, username, true)
			if err != nil {
				err = fmt.Errorf(errorMsg, err)
				log.Println(err.Error())
				return
			}
		}(targetUsername)
	}

	return nil
}

func (db *DataBase) UnBanUser(targetUsername string) error {

	errorMsg := "unban user error %w"

	var unbannedUserId int64

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(
		ctx,
		UNBAN_USER_BY_NAME,
		targetUsername,
	).Scan(&unbannedUserId)

	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	db.AddNotifGeneric(tx, unbannedUserId, 0, "unban", "You've been manually unbanned")

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		go func(username string) {
			_, err := db.GetUserByUserName(nil, username, true)
			if err != nil {
				err = fmt.Errorf(errorMsg, err)
				log.Println(err.Error())
				return
			}
		}(targetUsername)
	}

	return nil
}

func (db *DataBase) UnBanUser_ByTime() error {

	errorMsg := "automated unban user error %w"

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(
		ctx,
		UNBAN_USER_BY_TIME,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}
	defer rows.Close()

	unbannedUserIds := []int64{}

	for rows.Next() {
		var unbannedUserId int64
		if err := rows.Scan(&unbannedUserId); err != nil {
			return fmt.Errorf("failed to scan user ID: %w", err)
		}
		db.AddNotifGeneric(tx, unbannedUserId, 0, "unban", "Your ban has expired")

		if db.UseCache {
			_, err := db.GetUserById(tx, unbannedUserId)
			if err != nil {
				err = fmt.Errorf(errorMsg, err)
				log.Println(err.Error())
			}
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(errorMsg, err)
	}

	if db.UseCache {
		for _, id := range unbannedUserIds {
			go func(userId int64) {
				_, err := db.GetUserById(nil, userId)
				if err != nil {
					err = fmt.Errorf(errorMsg, err)
					log.Println(err.Error())
					return
				}
			}(id)
		}
	}

	return nil

}

func (db *DataBase) StartBanningSystem() {
	ticker := time.NewTicker(db.banRoutineInter.ToDuration())
	Wg.Add(1)

	go func() {
		defer Wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				db.UnBanUser_ByTime()
			case <-db.Ctx.Done():
				log.Println("Shuting down banning system routine")
				return
			}
		}
	}()
}
