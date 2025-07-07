CREATE TABLE Notifications (
    -- Not Null Values
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    ReceiverId INTEGER NOT NULL,
    SenderId INTEGER NOT NULL,
    Seen BOOLEAN DEFAULT FALSE,              -- Whether user has seen/read the notification
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Nullable
    NotifType TEXT,
    CommentId INTEGER, 
    UserReactionId INTEGER,
    SuperReportId INTEGER,
    BonusText TEXT,

    FOREIGN KEY(ReceiverId) REFERENCES User(id) ON DELETE CASCADE
    FOREIGN KEY (CommentId) REFERENCES Comment(id) ON DELETE CASCADE
    FOREIGN KEY (UserReactionId) REFERENCES UserReactions(Id) ON DELETE CASCADE
        CHECK (
        (UserReactionId IS NOT NULL AND CommentId IS NULL AND SuperReportId IS NULL) OR
        (CommentId IS NOT NULL AND UserReactionId IS NULL AND SuperReportId IS NULL) OR
        (SuperReportId IS NOT NULL AND CommentId IS NULL AND UserReactionId IS NULL) OR
        (SuperReportId IS NULL AND CommentId IS NULL AND UserReactionId IS NULL)
    )

);

-- Delete this if you need to allow multiple mod requests per user
CREATE UNIQUE INDEX one_modrequest_per_sender
ON Notifications(SenderId)
WHERE NotifType = 'modrequest';

-- adding efficiency when CountNotifs
CREATE INDEX idx_unseen_per_receiver
ON Notifications(ReceiverId)
WHERE Seen = 0;

CREATE INDEX idx_notifications_userid ON Notifications(ReceiverId);