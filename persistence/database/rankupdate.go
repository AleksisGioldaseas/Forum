package database

import (
	"log"
	"time"
)

func (db *DataBase) StartRankUpdater(cfg *RankingCfg) {

	ticker := time.NewTicker(cfg.Interval.ToDuration())
	Wg.Add(1)

	go func() {
		defer Wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-db.Ctx.Done():
				log.Println("Shuting down rank updater")
				return
			case <-ticker.C:
				db.rankUpdateRoutine(cfg)
			}
		}
	}()
}

func (db *DataBase) rankUpdateRoutine(cfg *RankingCfg) {
	log.Println("Rankupdate routine running")

	ctx, cancel, _ := db.newCtxTx(nil)
	defer cancel()
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		log.Println("rank update routine:", err)
		return
	}
	defer tx.Rollback()

	cutoff := time.Now().Add(-cfg.CutoffTime.ToDuration())

	rows, err := tx.QueryContext(ctx, GET_RECENT_POSTS, cutoff)
	if err != nil {
		log.Println("Error querying posts for rank update:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var created time.Time
		var likes, dislikes int64

		if err := rows.Scan(&id, &created, &likes, &dislikes); err != nil {
			log.Println("Error scanning post row:", err)
			continue
		}

		newScore := RankFunction(created, likes, dislikes, cfg)

		_, err := tx.ExecContext(ctx, UPDATE_RANKING, newScore, id)
		if err != nil {
			log.Println("Error updating RankScore for post ID", id, ":", err)
		}
	}

	if err := rows.Err(); err != nil {
		log.Println("Row iteration error:", err)
	}
}

func RankFunction(creationDate time.Time, likes, dislikes int64, cfg *RankingCfg) int64 {
	totalScore := cfg.DefaultRankScore
	totalScore += float64(likes) * cfg.LikeScore
	totalScore -= float64(dislikes) * cfg.DislikeScore
	totalScore -= totalScore / ((1.0 + time.Since(creationDate).Hours()/cfg.HalvingHourInterval) * cfg.TimePenalty)
	return int64(totalScore)
}
