package custom_errs

//contains error constants

import "errors"

// Generic errors
var (
	ErrInvalidArg = errors.New("invalid argument passed")
)

// Database operation errors
var (
	ErrDBConnetionIsLost      = errors.New("error database connection is lost")
	ErrNoRows                 = errors.New("warning: no rows affected")
	ErrNoRowsFound            = errors.New("warning: no rows found")
	ErrInvalidQuery           = errors.New("error invalid querry")
	ErrSettingSchema          = errors.New("error setting up schema")
	ErrIdNotFound             = errors.New("error: id not found in table")
	ErrNameNotFound           = errors.New("error: name not found in table")
	ErrWalCheckpoint          = errors.New("error during WAL checkpoint")
	ErrFetchingFromCache      = errors.New("error fetching from cache")
	ErrPlacingInCache         = errors.New("error placing entry in cache")
	ErrNilCache               = errors.New("error cache is of nil value")
	ErrInvalidReportArgs      = errors.New("commentId or postId must be 0")
	ErrExceededRowsLimit      = errors.New("rows call exceeds configured limit")
	ErrNullValueOnStructField = errors.New("null value on struct filed")
	ErrInteractionForbiden    = errors.New("interaction forbiden")
	ErrInvalidRole            = errors.New("invalid role")
)

// Authentication and user errors
var (
	ErrUsernameNotUnique    = errors.New("username already exists")
	ErrEmailNotUnique       = errors.New("email already exists")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrInvalidUserId        = errors.New("invalid user ID")
	ErrSessionNotFound      = errors.New("session token not found")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionCreation      = errors.New("failed to create session")
	ErrSessionAlreadyExists = errors.New("user already has an existing session")
	ErrInvalidSessionToken  = errors.New("invalid session token format")
	ErrUserNotFound         = errors.New("user not found")
	ErrUserCreationFailed   = errors.New("failed to create user")
	ErrUserUpdateFailed     = errors.New("failed to update user")
	ErrProfileNotFound      = errors.New("user profile not found")
)

// Content errors
var (
	ErrInvalidVoteAction   = errors.New("invalid vote action. Options: 'like' (1), 'dislike' (-1), 'neutral' (0)")
	ErrInvalidSortingArg   = errors.New("invalid sorting argument. Options: 'created', 'ranking', 'karma'")
	ErrUpdatingUserKarma   = errors.New("failed to update user karma")
	ErrDuplicateReaction   = errors.New("reaction already exists")
	ErrUnknownReactionInDb = errors.New("unknown reaction found in database")
	ErrCreatingCategory    = errors.New("failed to create category")
	ErrLinkPostToCategory  = errors.New("failed to link post to category")
	ErrInvalidDeleteStatus = errors.New("invalid delete status")
	ErrContentNotFound     = errors.New("requested content not found")
	ErrPermissionDenied    = errors.New("user lacks permission for this action")
	ErrInvalidTable        = errors.New("invalid table input")
	ErrImageTooBig         = errors.New("Image too big")
	ErrInvalidImageFile    = errors.New("Invalid image file: Only PNG, JPG, or GIF allowed")
)

// OAuth errors
var (
	ErrOAuthStateMismatch    = errors.New("OAuth state parameter mismatch")
	ErrOAuthCodeExchange     = errors.New("failed to exchange OAuth code for token")
	ErrOAuthUserInfoFetch    = errors.New("failed to fetch user info from provider")
	ErrOAuthProviderNotFound = errors.New("OAuth provider not supported")
	ErrOAuthEmailNotVerified = errors.New("OAuth email not verified")
	ErrOAuthAccountLinked    = errors.New("OAuth account already linked to another user")
	ErrOAuthNoEmail          = errors.New("OAuth provider didn't return an email")
	ErrOAuthTokenInvalid     = errors.New("invalid OAuth token")
	ErrOAuthMissingToken     = errors.New("missing OAuth access token")
	ErrOAuthInfoRequest      = errors.New("failed to create info request")
)

// security errors
var (
	ErrTooManyRequests   = errors.New("too many requests, please try again later")
	ErrCSRFTokenMismatch = errors.New("CSRF token validation failed")
	ErrInvalidInput      = errors.New("invalid input format")
	ErrOperationTimeout  = errors.New("operation timed out")
)

// Config errors
var (
	ErrNilConfigStruct = errors.New("nil configuration struct")
)

// Sse errors
var (
	ErrNotificationFailed = errors.New("notification failed")
	ErrUserNotConnected   = errors.New("user not connected")
	ErrGeneralSse         = errors.New("something went wrong")
)
