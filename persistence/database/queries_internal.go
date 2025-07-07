package database

/* This file contains a list of queries that are used by funcs internaly on db module.
They are not to be used as arguments in calls to funcs. If you need to make a func that accepts a query as argument
create a QueryName type and add it to queries map. Then define the func to accept QueryName type as argument instead of string*/

// SYSTEM -------------------------------------------
const (
	FOREIGN_KEYS_ENABLE = "PRAGMA foreign_keys = ON;"
	TRUNCATE            = `PRAGMA wal_checkpoint(TRUNCATE)`
	WAL_ALL             = `PRAGMA wal_checkpoint(FULL);`
	WAL_PRESSURE_CHECK  = `PRAGMA wal_checkpoint(PASSIVE);`
)

// TEST ----------------
const (
	GET_CATEGORIES_BY_POST_ID = `
	SELECT c.Name AS CategoryName
	FROM PostCategory pc
	JOIN Category c ON pc.CategoryId = c.Id
	WHERE pc.PostId = ?;`

	GET_ALL_CATEGORIES = `
	SELECT Name, Id
	FROM Category;`

	FUZZY_POST_TIME = `
    UPDATE Post
    SET Created = DATETIME('now', '-' || ABS(RANDOM() % 365) || ' days');`
)

// COMMON -----------------------------

const ()

// USER ------------------------------

const (
	ADD_USER = `
    INSERT INTO User 
    (UserName, Email, PasswordHash, ProfilePic, Bio, Role, Salt) 
    VALUES (?, ?, ?, ?, ?, ?, ?);`

	UPDATE_ROLE = `
    UPDATE User
    SET Role = ?
    WHERE UserName = ?
    RETURNING Id`

	ADD_OAUTH_USER = `
    INSERT INTO User 
    (UserName, Email, PasswordHash, ProfilePic, Bio, Role, OAuthSub, Salt) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	UPDATE_BIO_AND_PIC = `
    UPDATE User SET (Bio, ProfilePic) = (?, ?) WHERE Id = ?`

	UPDATE_BIO = `
    UPDATE User SET Bio = ? WHERE Id = ?`

	UPDATE_PIC = `
    UPDATE User SET ProfilePic = ? WHERE Id = ?`

	UPDATE_PASS = `
    UPDATE User
    SET PasswordHash = ?
    WHERE Id = ?`

	GET_USER_BY_EMAIL = `
    SELECT Id, UserName, Email, PasswordHash, ProfilePic, Bio, TotalKarma, Created, Role 
	FROM User 
	WHERE Email = ? AND Removed = 0 AND Deleted = 0`

	GET_USER_BY_SUB = `
    SELECT Id, UserName, Email, ProfilePic, OAuthSub
    FROM User
    WHERE OAuthSub = ?`

	UPDATE_USER_DATA = `
    UPDATE User
    SET 
    UserName = CASE WHEN COALESCE(?, '') <> '' THEN ? ELSE column1 END,
    Email = CASE WHEN COALESCE(?, '') <> '' THEN ? ELSE column2 END
    WHERE id = ?;
    `

	UPDATE_USER_KARMA_GEN = `UPDATE User
    SET TotalKarma = TotalKarma + ?
    WHERE Id = (SELECT UserId FROM %s WHERE Id = ?) AND Removed = 0 AND Deleted = 0;`

	COUNT_USERS = `
    SELECT COUNT(*) 
    FROM User
    WHERE Removed = 0 AND Deleted = 0`

	GET_USER_BY_ID = `
    SELECT Id, UserName, ProfilePic, Bio, TotalKarma, 
    Created, Role, Banned, BanExpDate, BanReason, BannedBy
    FROM User 
    WHERE Id = ? AND Removed = 0 AND Deleted = 0;`

	GET_USER_ID_BY_POST_ID = `
    SELECT UserId
    FROM Post
    WHERE Id = ?;`

	GET_USER_INFO_BY_UNAME = `
    SELECT Id, UserName, Email, PasswordHash, Salt
    FROM User
    WHERE LOWER(UserName) = LOWER(?) AND Removed = 0 AND Deleted = 0`

	GET_USER_BY_NAME = `
    SELECT Id, UserName, ProfilePic, Bio, TotalKarma, Created, Role, Banned, BanExpDate, BanReason, BannedBy
    FROM User
    WHERE LOWER(UserName) = LOWER(?) AND Removed = 0 AND Deleted = 0`

	BAN_USER_BY_NAME = `
    UPDATE User
    SET
        Banned = 1,
        BanExpDate = ?,
        BanReason = ?,
        BannedBy = ?
    WHERE UserName = ?
    RETURNING Id, Role`

	UNBAN_USER_BY_NAME = `
    UPDATE User
    SET
        Banned = 0
    WHERE UserName = ?
    RETURNING Id`

	UNBAN_USER_BY_TIME = `
    UPDATE User
    SET
        Banned = 0
    WHERE Banned = 1 AND BanExpDate <= ?
    RETURNING Id`

	UPDATE_USER_KARMA = `UPDATE User
	SET TotalKarma = (
    COALESCE((
        SELECT SUM(p.TotalKarma)
        FROM Post p
        WHERE p.UserId = User.Id
    ), 0)
    +
    COALESCE((
        SELECT SUM(c.TotalKarma)
        FROM Comment c
        WHERE c.UserId = User.Id
    ), 0)
	);`

	GET_OTHER_SESSIONS = `
    SELECT SessionToken 
    FROM UserSessions
	WHERE UserID = ? 
        AND SessionToken != ? 
        AND ExpiresAt > ?`
)

// POST ---------------------------------

const (
	CREATE_POST = `
    INSERT INTO Post 
    (UserId, Title, Body, 
    Img, Likes, Dislikes, 
    RankScore, Removed, 
    IsSuperReport, SuperReportCommentId, SuperReportPostId, SuperReportUserId) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

	TOGGLE_DELETE_STATUS = `
    UPDATE %s
    SET Deleted = ?, Body = ''
    WHERE Id = ? AND %s = ?`

	COUNT_POSTS = `
    SELECT COUNT(*) 
    FROM Post
    WHERE Removed = 0 AND Deleted = 0`

	//added scrubbing of body if
	GENERIC_POST_SEARCH = `
    SELECT 
    p.Id, 
    p.UserId,
    CASE 
        WHEN ? = 0 THEN  
            CASE 
                WHEN p.Removed = 0 THEN u.UserName  
                ELSE '(removed)' 
            END
        ELSE u.UserName  
    END AS username,
    p.Title,
    CASE 
        WHEN ? = 0 THEN  
            CASE 
                WHEN p.Removed = 0 THEN p.Body  
                ELSE '(removed)' 
            END
        ELSE p.Body  
    END AS body,
    p.Img, 
    p.likes, 
    p.dislikes,
    p.Created,
    ur.Reaction AS UserReaction,
    COALESCE(GROUP_CONCAT(c.Name, ', '), '') AS Categories,
    COALESCE(comment_count.total_comments, 0) AS CommentCount,
    p.Removed, 
    p.Deleted,
    p.Edited,
    u.Role,
    CASE 
        WHEN ? = 1 THEN COALESCE(GROUP_CONCAT(r.Message, '|!|!|'), '')
        ELSE ''
    END AS ReportMessages,
    CASE 
        WHEN ? = 1 THEN COALESCE(p.RemovalReason, '')
        ELSE ''
    END AS RemovalReason,
    CASE 
        WHEN ? = 1 THEN COALESCE(p.ModeratorName, '')
        ELSE ''
    END AS ModeratorName,
    p.IsSuperReport
    
    FROM Post p
    JOIN User u ON p.UserId = u.Id
    LEFT JOIN UserReactions ur ON p.Id = ur.PostId AND ur.UserId = ?
    LEFT JOIN PostCategory pc ON p.Id = pc.PostId
    LEFT JOIN Category c ON pc.CategoryId = c.Id
    LEFT JOIN (
        SELECT PostId, COUNT(*) AS total_comments
        FROM Comment
        GROUP BY PostId
    ) AS comment_count ON comment_count.PostId = p.Id
    LEFT JOIN Reports r ON r.PostId = p.Id
    %s
    %s -- Where clauses
    GROUP BY p.Id
    ORDER BY %s DESC
    LIMIT ? OFFSET ?
`

	//HERE
	GET_POST_BY_ID = `
    SELECT 
        p.Id,
        p.UserId,
        CASE
        WHEN p.Deleted = 1 THEN 
            '(deleted)'
        ELSE
            CASE 
            WHEN ? = 0 THEN  -- IF NOT MOD+
                CASE
                WHEN p.Removed = 1 THEN 
                    '(removed)'
                ELSE
                    u.UserName     
                END
            ELSE
                u.UserName  -- Fallback when parameter isn't 0
            END
        END AS username,
        p.Title, 
        CASE
        WHEN p.Deleted = 1 THEN 
            '(deleted)'
        ELSE
            CASE 
            WHEN ? = 0 THEN  -- IF NOT MOD+
                CASE
                WHEN p.Removed = 1 THEN 
                    '(removed)'
                ELSE
                    p.Body     
                END
            ELSE
                p.Body  -- Fallback when parameter isn't 0
            END
        END AS body,

        CASE 
        WHEN ? = 0 THEN  -- is viewer A Mod, if false (0) then scrub the image url
            CASE 
                WHEN p.Removed = 0 AND p.Deleted = 0 THEN p.Img 
                ELSE '' 
            END
        ELSE p.Img  -- Fallback when parameter isn't 0
        END AS Img,
        
        p.Likes,
        p.Dislikes, 
        p.RankScore,
        p.Created,
        p.TotalKarma, 
        ur.Reaction AS UserReactions,
        COALESCE(GROUP_CONCAT(c.Name, ', '), '') AS Categories,
        p.Removed, p.Deleted, p.Edited,
        p.IsSuperReport, p.SuperReportCommentId, p.SuperReportPostId, p.SuperReportUserId,
        CASE 
            WHEN ? = 1 THEN COALESCE(GROUP_CONCAT(r.Message, '|!|!|'), '')
            ELSE ''
        END AS ReportMessages,
        CASE 
            WHEN ? = 1 THEN COALESCE(p.RemovalReason, '')
            ELSE ''
        END AS RemovalReason,
        CASE 
            WHEN ? = 1 THEN COALESCE(p.ModeratorName, '')
            ELSE ''
        END AS ModeratorName,
        COALESCE(comment_count.total_comments, 0) AS CommentCount

    FROM Post p
    JOIN User u ON p.UserId = u.Id
    LEFT JOIN UserReactions ur ON p.Id = ur.PostId AND ur.UserId = ?
    LEFT JOIN PostCategory pc ON p.Id = pc.PostId
    LEFT JOIN Category c ON pc.CategoryId = c.Id
    LEFT JOIN Reports r ON r.PostId = p.Id
    LEFT JOIN (
        SELECT PostId, COUNT(*) AS total_comments
        FROM Comment
        GROUP BY PostId
    ) AS comment_count ON comment_count.PostId = p.Id
    WHERE p.Id = ? 
    GROUP BY p.Id;`

	SORT_POSTS_BY_CATEGORY = `
    SELECT 
        p.Id, p.UserId, u.UserName, p.Title, 
        p.Body, p.Img, p.Likes, p.Dislikes, 
        p.RankScore, p.Created,
        ur.Reaction AS UserReaction,
        COALESCE(GROUP_CONCAT(c.Name, ', '), '') AS Categories
    FROM Post p
    JOIN User u ON p.UserId = u.Id
    JOIN PostCategory pc ON p.Id = pc.PostId
    JOIN Category c ON pc.CategoryId = c.Id
    LEFT JOIN UserReactions ur ON p.Id = ur.PostId AND ur.UserId = ?
    WHERE c.Name IN (%s)
    GROUP BY p.Id
    ORDER BY ? DESC  
    LIMIT ? OFFSET ?;`

	UPDATE_POST = `
    UPDATE Post 
    SET Body = ?, Edited = 1
    WHERE Id = ? AND UserId = ?; `

	GET_RECENT_POSTS = `
    SELECT Id, Created, likes, dislikes
    FROM Post
    WHERE Created > ? AND Removed = 0 AND Deleted = 0`

	UPDATE_RANKING = `
    UPDATE Post
    SET RankScore = ?
    WHERE Id = ?`
)

// MODERATOR ----------------------------------------------

const (
	TOGGLE_REMOVE_STATUS = `
    UPDATE %s
    SET Removed = ?, RemovalReason = ?, ModeratorName = ?, RemovedTime = ?
    WHERE Id = ?`

	NEW_MOD_LOG = `
    INSERT INTO ModeratorLog
    (ActionType, ModeratorId, TableName, RowId, Body)
    VALUES (?, ?, ?, ?, ?)`
)

// ADMIN --------------------------------------------------

const (
	GET_ADMIN_IDS = `
	SELECT id
	FROM User
	WHERE Role = 3
	`
)

// CATEGORIES ---------------------------------------------

const (
	CREATE_CAT_OR_RETURN_ID = `
    INSERT INTO Category (Name) VALUES (?)
    ON CONFLICT(Name) DO UPDATE SET Name=excluded.Name
    RETURNING Id;`

	LINK_POST_CATEGORY = `
    INSERT INTO PostCategory 
    (PostId, CategoryId)
    VALUES (?, ?);`

	DELETE_CATEGORY = `
    DELETE FROM Category WHERE Name = ?`
)

// REACTIONS ON CONTENT

const (
	UPDATE_REACTION_ON_CNT_AND_RETURN_TXT = `
    UPDATE %s
    SET %s = %s + 1 
    WHERE Id = ?
    RETURNING SUBSTR(%s, 1, 40);`

	TOGGLE_REACTION_ON_CNT = `
    UPDATE %s
    SET likes = likes + ?,
    dislikes = dislikes - ? 
    WHERE Id = ?;`

	UNDO_REACTION = `
    UPDATE %s
    SET %s = %s - 1
    WHERE Id = ?;`

	UNDO_LIKE = `
    UPDATE %s
    SET Likes = Likes - 1
    WHERE Id = ?;`

	UNDO_DISLIKE = `
    UPDATE %s
    SET Dislikes = Dislikes - 1
    WHERE Id = ?;`
)

// USER REACTIONS --------------------------------

const (
	NEW_USER_REACTIONS_GEN = `
    INSERT INTO UserReactions (Reaction, %s, UserId)
    VALUES (?, ?, ?)
    `

	TOGGLE_USER_REACTIONS = `
    UPDATE UserReactions
    SET Reaction = 
    CASE
        WHEN UserReactions.Reaction = 1 THEN -1
        WHEN UserReactions.Reaction = -1 THEN 1
    END
    WHERE %s = ? AND UserId = ?`

	UPDATE_USER_REACTION = `
    UPDATE UserReactions
	SET Reaction = ?
	WHERE UserId = ? AND %s = ?`

	DELETE_REACTION = `
    DELETE FROM UserReactions
    WHERE %s = ? AND UserId = ?
    RETURNING Id;`

	GET_FROM_USER_REACTIONS_GEN = `
    SELECT Reaction
    FROM UserReactions
    WHERE %s = ? AND UserId = ?;`

	GET_USER_REACTION = `
	SELECT r.UserId, r.PostId, r.CommentId, r.Reaction
	FROM UserReactions r
	WHERE r.Id = ?`
)

// COMMENTS --------------------------------------------------

const (
	CREATE_COMMENT = `
    INSERT INTO Comment
    (UserId, PostID, Body)
    VALUES (?, ?, ?);`

	COUNT_COMMENTS = `
    SELECT COUNT(*) 
    FROM Comment
    WHERE PostId = ?`

	UPDATE_COMMENT = `
    UPDATE Comment 
    SET Body = ?, Edited = 1
    WHERE Id = ? AND UserId = ?;`

	GET_COMMENTS = `
    SELECT 
        c.Id, c.UserId, c.PostId,

        CASE 
        WHEN c.Deleted = 1 THEN 
            '(deleted)'
        ELSE
            CASE
            WHEN c.Removed = 1 AND ? < 2 THEN 
                '(removed)'
            ELSE 
                c.Body
            END
        END AS Body,

        c.Created,

        CASE 
        WHEN c.Deleted = 1 THEN 
            '(deleted)'
        ELSE
            CASE
            WHEN c.Removed = 1 AND ? < 2 THEN 
                '(removed)'
            ELSE 
                u.UserName
            END
        END AS UserName,

        c.Removed, c.Deleted, c.Edited, u.Role, c.likes, c.dislikes, c.TotalKarma,
        ur.Reaction,

        CASE 
            WHEN ? >= 2 THEN COALESCE(GROUP_CONCAT(r.Message, '|!|!|'), '')
            ELSE ''
        END AS ReportMessages,

        CASE 
            WHEN ? >= 2 THEN COALESCE(c.RemovalReason, '')
            ELSE ''
        END AS RemovalReason,

        CASE 
            WHEN ? >= 2 THEN COALESCE(c.ModeratorName, '')
            ELSE ''
        END AS ModeratorName

    FROM Comment c
    JOIN User u ON c.UserId = u.Id
    LEFT JOIN UserReactions ur ON ur.UserId = ? AND ur.CommentId = c.Id
    LEFT JOIN Reports r ON r.CommentId = c.Id

    WHERE 
        (c.PostId = ? OR 1 = ? OR 1 = ?)
        %s
    GROUP BY c.Id
    ORDER BY %s
    LIMIT ? OFFSET ?;
    `

	GET_COMMENT = `
    SELECT c.Id, c.UserId, c.PostId, c.Body, c.Created, 
    u.UserName, c.Removed, c.Deleted, u.Role, c.likes, c.dislikes, c.TotalKarma,
    CASE 
        WHEN ? >= 2 THEN COALESCE(GROUP_CONCAT(r.Message, '|!|!|'), '')
        ELSE ''
    END AS ReportMessages,
    CASE 
        WHEN ? >= 2 THEN COALESCE(c.RemovalReason, '')
        ELSE ''
    END AS RemovalReason,
    CASE 
        WHEN ? >= 2 THEN COALESCE(c.ModeratorName, '')
        ELSE ''
    END AS ModeratorName
    
    FROM Comment c
    JOIN User u ON c.UserId = u.Id
    LEFT JOIN Reports r ON r.CommentId = c.Id
    WHERE c.Id = ?
    GROUP BY c.Id
    `
)

// LOG IN CHECKS -------------------------------
const (
	CHECK_USERNAME_UNIQUE = `
    SELECT 1 
    FROM User 
    WHERE UserName = ? 
    LIMIT 1;`

	CHECK_EMAIL_UNIQUE = `
    SELECT 1 
    FROM User 
    WHERE Email = ? 
    LIMIT 1;`
)

// SESSIONS ----------------------------------

const (
	GET_USER_FROM_SESSION = `
    SELECT 
        s.ExpiresAt,
        u.Id,
        u.UserName,
        u.Email,
        u.ProfilePic,
        u.Bio,
        u.TotalKarma,
        u.Created,
        u.Role,
        u.Banned,
        u.BanExpDate,
		u.BanReason,
		u.BannedBy
    FROM UserSessions s
    JOIN User u ON s.UserID = u.Id
    WHERE s.SessionToken = ? AND u.Removed = 0 AND u.Deleted = 0;`

	STORE_SESSION = `
    INSERT INTO UserSessions (SessionToken, UserId, ExpiresAt) 
    VALUES (?, ?, ?)`

	DELETE_SESSION = `
    DELETE FROM UserSessions 
    WHERE SessionToken = ?`

	HAS_SESSION = `
    SELECT COUNT(*) 
    FROM UserSessions 
    WHERE UserID = ? AND ExpiresAt > ?`

	CLEAN_UP = `
    DELETE FROM UserSessions 
    WHERE ExpiresAt < DATETIME('now')`
)

// REPORTS -----------------------------------

const (
	ADD_REPORT = `
    INSERT INTO Reports
    (SenderId, %s, Message)
    VALUES (?, ?, ?)`

	GET_REPORTS = `
    SELECT 
    Id, SenderId, PostId, CommentId, Message, Created
    FROM Reports
    WHERE %s = ?
    `
)

// USER ACTIVITY -----------------------------

const (
	INSERT_USER_ACTIVITY = `
    INSERT INTO UserActivity 
    (UserID, PostId, CommentId, ReactionId) 
    VAlUES (?, ?, ?, ?)`

	GET_USER_ACTIVITIES = `
    SELECT 
        ua.Id,
        ua.UserId,
        ua.PostId,
        ua.CommentId,
        ua.ReactionId,
        ua.Created
    FROM UserActivity ua
	WHERE ua.UserId = ?
	ORDER BY ua.Created DESC
	LIMIT ? OFFSET ?;`

	GET_USER_ACTIVITIES_BY_UNAME = `
    SELECT 
        ua.Id,
        ua.UserId,
        ua.PostId,
        ua.CommentId,
        ua.ReactionId,
        ua.Created
    FROM UserActivity ua
    JOIN User u ON u.Id = ua.UserId 
	WHERE u.UserName = ?
	ORDER BY ua.Created DESC
	LIMIT ? OFFSET ?;`

	DELETE_USER_ACTIVITY = `
    DELETE FROM UserActivity WHERE userId = ? AND %s = ?`

	GET_POST_ACTIVITY = `
    SELECT p.Title, SUBSTR(p.Body, 1, 40) AS Body, p.Removed, p.Deleted
	FROM Post p
	WHERE ID = ?`

	GET_COM_ACTIVITY = `
	SELECT c.PostId, p.Title, SUBSTR(c.Body, 1, 40) AS Body, p.Removed, p.Deleted
	FROM Comment c
	LEFT JOIN Post p ON p.Id = c.PostId
	WHERE c.Id = ?;`
)

// NOTIFICATIONS ------------------------------------------------

const (
	GET_NOTIFS = `
    SELECT
        -- not nulls
        ntf.ReceiverId,
        ntf.SenderId,
        su.UserName,
        ntf.Seen,
        ntf.Created,
        ntf.NotifType,

        -- Super Post
        super.Id,
        super.Title,
        
        -- reactions
        rea.Reaction,
        
        -- comment reacted
        reacom.Id,
        reacom.PostId,
        parentPost.Title,

        -- post reacted
        reapst.Id,
        reapst.Title,
        
        -- comment on post
        com.Id,
        com.Body,

        -- post that was commented
        compst.Id,
        compst.Title,
        compst.IsSuperReport,

        ntf.BonusText,
        
        COUNT(*) OVER() AS TotalNotificationCount

    FROM Notifications ntf

    LEFT JOIN User su ON ntf.SenderId = su.Id
    -- Join on reaction
    LEFT JOIN UserReactions rea ON rea.Id = ntf.UserReactionId
        LEFT JOIN Comment reacom 
            ON reacom.Id = rea.CommentId 
            AND reacom.deleted = 0
        LEFT JOIN Post parentPost
            ON reacom.PostId = parentPost.ID
        LEFT JOIN Post reapst 
            ON reapst.Id = rea.PostId 
            AND reapst.deleted = 0

    LEFT JOIN Comment com 
        ON com.Id = ntf.CommentId 
        AND com.removed = 0 AND com.deleted = 0
        LEFT JOIN Post compst 
            ON compst.Id = com.PostId 
            AND compst.deleted = 0

    LEFT JOIN Post super 
        ON super.Id = ntf.SuperReportId 
        AND super.removed = 0 AND super.deleted = 0

    WHERE ReceiverId = ?
    ORDER BY ntf.Created DESC
    LIMIT ? OFFSET ?;
    `
	INSERT_NOTIF = `
	INSERT INTO Notifications (ReceiverId, SenderId, NotifType, BonusText) VALUES (?, ?, ?, ?)`

	ADD_NOTIF_SUPER = `
	INSERT INTO Notifications 
	(ReceiverId, SenderId, SuperReportId) 
	VALUES (?, ?, ?) 
	`

	ADD_NOTIF_COMMENT = `
    INSERT INTO Notifications (commentId, SenderId, ReceiverId)
    SELECT
        c.Id AS commentId,
        c.UserId AS SenderId,
        p.UserId AS ReceiverId
    FROM Comment c
    JOIN Post p ON p.Id = ?
    WHERE c.Id = ? AND c.UserId != p.UserId
    RETURNING ReceiverId;
    `

	ADD_NOTIF_REACTION = `
    INSERT INTO Notifications (UserReactionId, SenderId, ReceiverId)
    SELECT
        ur.Id AS UserReactionId,
        ur.UserId AS SenderId,
        COALESCE(p.UserId, c.UserId) AS ReceiverId
    FROM UserReactions ur
    LEFT JOIN Post p ON ur.PostId = p.Id
    LEFT JOIN Comment c ON ur.CommentId = c.Id
    WHERE ur.Id = ? AND SenderId != ReceiverId
    RETURNING ReceiverId;
    `

	DELETE_NOTIFICATION = `
	DELETE FROM Notifications
	WHERE SenderUserName = (SELECT UserName FROM User WHERE Id = ?) AND TargetType = ? AND TargetId = ?`

	HAS_UNSEEN = `
    SELECT HasUnseenNotifications
    FROM User
    WHERE Id = ?;`

	NOTIFY_USER = `
    UPDATE User
    Set HasUnseenNotifications = true
    WHERE Id = ?;`

	MARK_ALL_SEEN = `
    UPDATE Notifications
    SET Seen = true
    WHERE ReceiverId = ? AND Seen = false;
    `
	SEEN_FALSE = `
    UPDATE User
    Set HasUnseenNotifications = false
    WHERE Id = ?;`

	COUNT_NOTIFS = `
    SELECT COUNT(*) 
    FROM Notifications
    WHERE ReceiverId = ? AND Seen = 0`
)

// IMAGES --------------------------------------------------------------

const (
	INSERT_IMAGE = `
    INSERT INTO Images (FileName) VALUES (?)`

	IS_IMAGE_HIDDEN = `
    SELECT Hide FROM Images WHERE FileName = ?`

	TOGGLE_HIDE_IMAGE = `
	UPDATE Images
	SET Hide = ?
	WHERE FileName = (
	    SELECT %s FROM %s WHERE Id = ?
    )`
)
