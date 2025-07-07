package database

import (
	"context"
	"database/sql"
	"forum/persistence/cache"
	"forum/utils"
	"sync"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// High level struct used on call to open database
type DBConfig struct {
	Ctx             context.Context
	Path            []string            `json:"path"`
	Wal             Wal                 `json:"wal"`
	UseCache        bool                `json:"use_cache"`
	Cache           CacheSetUp          `json:"cache_setup"`
	CleanUpSessions utils.Duration      `json:"clean_up_sessions"`
	RankCfg         RankingCfg          `json:"ranking_cfg"`
	Limits          Limits              `json:"limits"`
	BanRoutineInter utils.Duration      `json:"ban_routine_interval"`
	SystemImages    map[string]struct{} `json:"system_images"` // List of non removable images
}

type CacheSetUp struct {
	UsersCacheLimit      int `json:"users_cache_limit"`
	PostsCacheLimit      int `json:"posts_cache_limit"`
	CategoriesCacheLimit int `json:"categories_cache_limit"`
}

type RankingCfg struct {
	DefaultRankScore    float64        `json:"default_rank_score"` // all posts begin with a base score, so that they show up over older posts of similar score
	LikeScore           float64        `json:"like_score"`
	DislikeScore        float64        `json:"dislike_score"`         // dislike score has less weight so that controverial posts hold on a little more
	TimePenalty         float64        `json:"time_penalty"`          // extra rank penalty multiplier
	HalvingHourInterval float64        `json:"halving_hour_interval"` // how many hours it takes for the score to halve
	CutoffTime          utils.Duration `json:"cutoff_time"`           // cutoff time of creation of posts to look up
	Interval            utils.Duration `json:"interval"`              // the interval that the ranking proccess takes place
}

type Limits struct {
	RowsLimit      int `json:"rows_limit"`
	MaxUsername    int `json:"max_username"`
	MinUsername    int `json:"min_username"`
	MaxPass        int `json:"max_pass"`
	MinPass        int `json:"min_pass"`
	MaxBio         int `json:"max_bio"`
	MaxTitle       int `json:"max_title"`
	MinTitle       int `json:"min_title"`
	MaxCommentBody int `json:"max_comment_body"`
	MaxPostBody    int `json:"max_post_body"`
	MinBody        int `json:"min_body"`
	MaxCategories  int `json:"max_categories"`
	MaxReportBody  int `json:"max_report_body"`
}

type Wal struct {
	AutoTruncate     bool           `json:"auto_truncate"`
	TruncateInterval utils.Duration `json:"truncate_interval"`
	CacheSize        string         `json:"cache_size"`
	Synchronous      string         `json:"synchronous"`
}

type DataBase struct {
	Ctx             context.Context
	conn            *sql.DB
	cache           *cache.Cache
	mu              sync.Mutex
	cacheCfg        *cache.CHConfig
	UseCache        bool
	Limits          *Limits
	systemImages    *map[string]struct{}
	banRoutineInter utils.Duration
}

func NewDBConfig() *DBConfig {
	return &DBConfig{}
}
