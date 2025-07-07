package database

import (
	"database/sql"
	"fmt"
	"forum/persistence/cache"
	"forum/utils"
	"log"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/* Entry point*/

// Establishes connection with database and returns DataBase type instance that implements all database funcs.
// Always returns a connection even if errors occur unless error is dbErr.ErrOpeningDB in which case connection is nil.
// Also spawns Truncate() go routine that needs to be handled using the Wg defined with it
func Open(cfg *DBConfig) (db *DataBase, err error) {
	fmt.Println("*** DATABASE ***")
	fmt.Println("- TruncateInterval:", cfg.Wal.TruncateInterval.String())
	fmt.Println("- Clean Up Sessions Interval:", cfg.CleanUpSessions.String())
	fmt.Println("- Ban Routine Interval:", cfg.BanRoutineInter)
	fmt.Println("- Rows Limit Per Call:", cfg.Limits.RowsLimit)
	fmt.Println("- Synchronous:", cfg.Wal.Synchronous)
	fmt.Println("- Cache Size:", cfg.Wal.CacheSize)
	log.Println("Opening Database...")
	db = NewDb(cfg)
	db.conn, err = sql.Open("sqlite3", filepath.Join(cfg.Path...))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return db, err
	}

	db.walSetUp(cfg.Wal.Synchronous, cfg.Wal.CacheSize)

	// Enable foreign keys for older sql versions
	_, err = db.conn.Exec(FOREIGN_KEYS_ENABLE)
	if err != nil {
		log.Println(err)
	}

	// Spawning auto truncate
	if cfg.Wal.AutoTruncate {
		if cfg.Wal.TruncateInterval == 0 {
			cfg.Wal.TruncateInterval = utils.Duration(1 * time.Hour)
		}

		db.StartTruncateRoutine(cfg.Wal.TruncateInterval)
	}

	// Setting up cache
	if cfg.UseCache {
		db.cacheCfg = &cache.CHConfig{
			UsersLimit:      cfg.Cache.UsersCacheLimit,
			PostsLimit:      cfg.Cache.PostsCacheLimit,
			CategoriesLimit: cfg.Cache.CategoriesCacheLimit,
		}
		db.cache = cache.NewCache(db.cacheCfg)
		err = db.loadCache()
	}

	// Spawning clean up sessions
	if cfg.CleanUpSessions == 0 {
		cfg.CleanUpSessions = utils.Duration(10 * time.Minute)
	}

	db.sessionCleanupTask(cfg.CleanUpSessions)

	db.StartRankUpdater(&cfg.RankCfg)

	db.StartBanningSystem()

	return db, err
}

func (db *DataBase) Ping() error {
	if err := db.conn.Ping(); err != nil {
		return err
	}
	log.Println("Database connected successfully.")
	return nil
}

func (db *DataBase) Close() {
	if db != nil {
		log.Println("Closing database...")
		db.conn.Close()
	}
}

func NewDb(config *DBConfig) *DataBase {
	return &DataBase{
		Ctx:             config.Ctx,
		UseCache:        config.UseCache,
		Limits:          &config.Limits,
		systemImages:    &config.SystemImages,
		banRoutineInter: config.BanRoutineInter,
	}
}
