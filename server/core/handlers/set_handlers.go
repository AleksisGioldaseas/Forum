package handlers

//sets all the handlers to their respective urls

import (
	"fmt"
	"forum/persistence/database"
	"forum/server/core/config"
	"forum/utils"
	"net/http"
	"path"
)

const (
	GUEST = 0
	USER  = 1
	MOD   = 2
	ADMIN = 3
)

var (
	ALLOWBANNED = true
	BLOCKBANNED = false
)

type Config struct {
	Images                *imgConfig                `json:"image"`
	MaxPostSize           utils.FileSize            `json:"max_post_size"`
	XorKey                string                    `json:"xor_key"`
	CookieExpirationHours utils.Duration            `json:"cookie_expiration_hours"`
	RateLimits            map[string]map[string]any `json:"rate_limits"`
}

var Configuration *Config

// This is where all our handlers are defined

type endpoint struct {
	path                     string
	rateLimitMaxRequests     int     //how many requests can you do...
	ratelimitSecondsInterval float64 //in this time interval, in seconds
	requiredRole             int     // what role user that reaches endpoint needs to have to interact with it
	nextHandler              func(http.ResponseWriter, *http.Request, *database.DataBase, *config.User)
	allowBanned              bool
}

func makeEndpoint(path string, rateLimitMaxRequests int, ratelimitSecondsInterval float64, requiredRole int, nextHandler func(http.ResponseWriter, *http.Request, *database.DataBase, *config.User), allowBanned bool) endpoint {
	return endpoint{path: path, rateLimitMaxRequests: rateLimitMaxRequests, ratelimitSecondsInterval: ratelimitSecondsInterval, requiredRole: requiredRole, nextHandler: nextHandler, allowBanned: allowBanned}
}

func SetHandlers(db *database.DataBase) *http.ServeMux {

	go syncMapCleaner()

	mux := http.NewServeMux()

	// Build the path to the static directory
	staticDirCss := path.Join("web", "static")
	staticDirImg := path.Join("web", "images")
	staticDirJs := path.Join("web", "js")

	// Serve static files at /static path
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDirCss))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(staticDirImg))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(staticDirJs))))

	endpoints := []endpoint{
		// ========== APIS ===========
		makeEndpoint("/image/", 40, 20, GUEST, serveImageAPI, ALLOWBANNED),
		makeEndpoint("/useractivity/", 40, 20, GUEST, userActivityAPI, ALLOWBANNED),

		// ========== PAGE HANDLERS ==========
		// Basic pages
		makeEndpoint("/", 20, 20, GUEST, HomePageHandler, ALLOWBANNED),
		makeEndpoint("/profile/", 20, 20, GUEST, ProfilePageHandler, ALLOWBANNED),
		makeEndpoint("/post/", 20, 20, GUEST, PostPageHandler, ALLOWBANNED),

		// User-facing form pages
		makeEndpoint("/postform", 5, 20, USER, CreatePostPageHandler, BLOCKBANNED),
		makeEndpoint("/notificationfeed", 10, 20, USER, NotificationsPageHandler, ALLOWBANNED),

		// Moderation pages
		makeEndpoint("/superreportform", 100, 20, MOD, SuperReportPageFormHandler, BLOCKBANNED),
		makeEndpoint("/superreport/", 100, 20, MOD, SuperReportPageHandler, BLOCKBANNED),
		makeEndpoint("/removedposts", 100, 20, MOD, RemovedPostsPageHandler, BLOCKBANNED),
		makeEndpoint("/removedcomments", 100, 20, MOD, RemovedCommentsPageHandler, BLOCKBANNED),
		makeEndpoint("/reportedposts", 100, 20, MOD, ReportedPostsPageHandler, BLOCKBANNED),
		makeEndpoint("/reportedcomments", 100, 20, MOD, ReportedCommentsPageHandler, BLOCKBANNED),

		// Admin pages
		makeEndpoint("/superreports", 100, 20, ADMIN, SuperReportsViewPageHandler, BLOCKBANNED),
		makeEndpoint("/editcategories", 100, 20, ADMIN, EditCategoriesPageHandler, BLOCKBANNED),

		// ========== EVENT HANDLERS ==========
		// Authentication
		makeEndpoint("/login", 5, 30, GUEST, LoginHandler, ALLOWBANNED),
		makeEndpoint("/logout", 5, 30, USER, LogoutHandler, ALLOWBANNED),
		makeEndpoint("/logoutall", 5, 30, USER, LogoutAllHandler, ALLOWBANNED),
		makeEndpoint("/signup", 5, 30, GUEST, SignupHandler, ALLOWBANNED),
		makeEndpoint("/auth/google", 10, 25, GUEST, GoogleLoginHandler, ALLOWBANNED),
		makeEndpoint("/auth/google/callback", 10, 25, GUEST, GoogleCallbackHandler, ALLOWBANNED),
		makeEndpoint("/auth/github", 10, 25, GUEST, GithubLoginHandler, ALLOWBANNED),
		makeEndpoint("/auth/github/callback", 10, 25, GUEST, GithubCallbackHandler, ALLOWBANNED),

		// Content creation
		makeEndpoint("/createpost", 5, 120, USER, CreatePostHandler, BLOCKBANNED),
		makeEndpoint("/createcomment", 5, 20, USER, CreateCommentHandler, BLOCKBANNED),
		makeEndpoint("/createsuperreport", 100, 20, MOD, CreateSuperReportHandler, BLOCKBANNED),

		// Content modification
		makeEndpoint("/postedit", 5, 20, USER, EditPostHandler, BLOCKBANNED),
		makeEndpoint("/commentedit", 5, 20, USER, EditCommentHandler, BLOCKBANNED),
		makeEndpoint("/profilepic", 5, 2000, USER, ProfilePicHandler, BLOCKBANNED),
		makeEndpoint("/updatebio", 5, 20, USER, BioHandler, BLOCKBANNED),

		// Content deletion
		makeEndpoint("/postdelete", 10, 20, USER, DeletePostHandler, BLOCKBANNED),
		makeEndpoint("/commentdelete", 10, 20, USER, DeleteCommentHandler, ALLOWBANNED),
		makeEndpoint("/postremove", 100, 20, MOD, RemovePostHandler, BLOCKBANNED),
		makeEndpoint("/commentremove", 100, 20, MOD, RemoveCommentHandler, BLOCKBANNED),

		// Content approval
		makeEndpoint("/postapprove", 100, 20, MOD, ApprovePostHandler, BLOCKBANNED),
		makeEndpoint("/commentapprove", 100, 20, MOD, ApproveCommentHandler, BLOCKBANNED),

		// Voting
		makeEndpoint("/votepost", 15, 20, USER, VotePostHandler, ALLOWBANNED),
		makeEndpoint("/votecomment", 15, 20, USER, VoteCommentHandler, ALLOWBANNED),

		// Reporting
		makeEndpoint("/postreport", 5, 20, USER, ReportPostHandler, BLOCKBANNED),
		makeEndpoint("/commentreport", 5, 20, USER, ReportCommentHandler, BLOCKBANNED),

		// Moderation requests
		makeEndpoint("/modrequest", 5, 20, USER, ModRequestHandler, BLOCKBANNED),

		// Data listing
		makeEndpoint("/postlist", 20, 20, GUEST, GetPostListHandler, ALLOWBANNED),
		makeEndpoint("/commentlist", 20, 20, GUEST, GetCommentListHandler, ALLOWBANNED),
		makeEndpoint("/notificationlist", 20, 20, USER, GetNotificationListHandler, ALLOWBANNED),

		// Notifications
		makeEndpoint("/ssenotifications", 10, 20, USER, NotificationSSEHandler, ALLOWBANNED),
		makeEndpoint("/notificationseen", 10, 20, USER, NotificationsSeenHandler, ALLOWBANNED),

		// Mod actions
		makeEndpoint("/banuser", 100, 20, MOD, BanUserHandler, BLOCKBANNED),
		makeEndpoint("/unbanuser", 100, 20, MOD, UnBanUserHandler, BLOCKBANNED),

		// Admin actions
		makeEndpoint("/demotemoderator", 100, 20, ADMIN, DemoteModeratorHandler, BLOCKBANNED),
		makeEndpoint("/promoteuser", 100, 20, ADMIN, PromoteUserHandler, BLOCKBANNED),
		makeEndpoint("/removecategory", 100, 20, ADMIN, RemoveCategoryHandler, BLOCKBANNED),
		makeEndpoint("/addcategory", 100, 20, ADMIN, AddCategoryHandler, BLOCKBANNED),
	}

	for _, endpoint := range endpoints {
		//set default values given from go's source code
		limitCount := endpoint.rateLimitMaxRequests
		limitInterval := endpoint.ratelimitSecondsInterval

		//check configuration maps for overriding rate limit values
		if limits, ok := Configuration.RateLimits[endpoint.path]; ok {
			lCount, ok1 := limits["rate_limit_count"].(float64)
			lSeconds, ok2 := limits["rate_limit_second_interval"].(float64)
			if !ok1 || !ok2 {
				fmt.Println("something wrong with rate limit key in configs, suspected bad configs format: ", lCount, ok1, lSeconds, ok2)
			} else {
				limitCount, limitInterval = int(lCount), lSeconds
			}
		}

		universalCount, ok1 := Configuration.RateLimits["universal"]["rate_limit_count"].(float64)
		universalSeconds, ok2 := Configuration.RateLimits["universal"]["rate_limit_second_interval"].(float64)
		if !ok1 || !ok2 {
			fmt.Println("something wrong with rate limit key in configs, suspected bad configs format: ", universalCount, ok1, universalSeconds, ok2)
			universalCount = 30
			universalSeconds = 2
		}

		mux.HandleFunc(endpoint.path, func(w http.ResponseWriter, r *http.Request) {
			AuthMiddleware(endpoint.requiredRole, w, r, db, endpoint.nextHandler, limitCount, limitInterval, int(universalCount), universalSeconds, endpoint.allowBanned)
		})
	}

	return mux
}
