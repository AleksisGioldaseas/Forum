package database

const (
	// TOGGLE
	REMOVE  = 1
	RESTORE = 0

	// USERS
	GUEST  = 0
	USER   = 1
	MOD    = 2
	ADMIN  = 3
	SYSTEM = -1

	// INT BOOLS
	TRUE  = 1
	FALSE = 0

	// MESSAGES
	REMOVED_OR_DELETED_CONTENT = "Content was removed or deleted"
)

var TABLES = map[string]struct{}{
	"user":    {},
	"comment": {},
	"post":    {},
}
