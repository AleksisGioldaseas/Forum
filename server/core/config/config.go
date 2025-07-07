package config

//contains settings for the server, variables that are expected to be changed during development or production should go here.

import "time"

type User struct {
	ID           int64     `json:"id"`
	UserName     string    `json:"user_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // This excludes it from JSON serialization!
	Salt         string    `json:"-"`
	ProfilePic   *string   `json:"profile_pic"`
	Description  string    `json:"description"`
	Bio          *string   `json:"bio"`
	TotalKarma   int64     `json:"total_karma"`
	Created      time.Time `json:"created"`
	Role         int       `json:"role"`
	OAuthSub     string    `json:"-"`
	Removed      int       `json:"removed"`
	Deleted      int       `json:"deleted"`
	Banned       int       `json:"banned"`
	BanExpDate   time.Time `json:"banexpdate"`
	BanReason    string    `json:"ban_reason"`
	BannedBy     string    `json:"banned_by"`
}

// Data format for all post info in the database
type Post struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	UserName     string    `json:"user_name"`
	UserImg      string    `json:"user_img"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	PostImg      string    `json:"post_img"`
	CreationDate time.Time `json:"creation_date"`
	Likes        uint64    `json:"likes"`
	Dislikes     uint64    `json:"dislikes"`
	RankScore    int64     `json:"rank_score"`
	TotalKarma   int64     `json:"total_karma"`
	Categories   []string  `json:"categories"`
	Edited       int       `json:"edited"`

	Removed       int    `json:"removed"`
	RemovalReason string `json:"removal_reason"`
	ModeratorName string `json:"mod_name"`

	Deleted              int      `json:"deleted"`
	CommentCount         int      `json:"comment_count"`
	IsSuperReport        bool     `json:"is_super_report"`
	SuperReportCommentId int64    `json:"super_report_comment_id"`
	SuperReportPostId    int64    `json:"super_report_post_id"`
	SuperReportUserId    int64    `json:"super_report_user_id"`
	UserRole             int      `json:"user_role"`
	UserReaction         *int     `json:"user_reaction"`
	Reports              []string `json:"reports"`
}

type Comment struct {
	ID           int64     `json:"id"`
	UserName     string    `json:"user_name"`
	PostID       int64     `json:"post_id"`
	UserID       int64     `json:"user_id"`
	Body         string    `json:"body"`
	CreationDate time.Time `json:"creation_date"`
	Edited       int

	Removed       int    `json:"removed"`
	RemovalReason string `json:"removal_reason"`
	ModeratorName string `json:"mod_name"`

	Deleted      int      `json:"deleted"`
	Likes        uint64   `json:"likes"`
	Dislikes     uint64   `json:"dislikes"`
	TotalKarma   int64    `json:"total_karma"`
	UserRole     int      `json:"user_role"`
	UserReaction *int     `json:"user_reaction"`
	Reports      []string `json:"reports"`
}

type Certifications struct {
	UseHTTPS bool     `json:"use_https"`
	File     []string `json:"file"`
	Key      []string `json:"key"`
}

type Vote struct {
	Id     int
	Action string //"like", "dislike", "neutral"
}

func NewUser() *User {
	return &User{}
}

func NewPost() *Post {
	return &Post{}
}

func NewComment() *Comment {
	return &Comment{}
}
