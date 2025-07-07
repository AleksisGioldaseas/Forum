CREATE TABLE UserActivity (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    UserId INTEGER NOT NULL,
    PostId INTEGER,
    CommentId INTEGER,
    ReactionId INTEGER,
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(UserId) REFERENCES User(id) ON DELETE CASCADE
    FOREIGN KEY (PostId) REFERENCES Post(id) ON DELETE CASCADE
    FOREIGN KEY (CommentId) REFERENCES Comment(id) ON DELETE CASCADE
    FOREIGN KEY (ReactionId) REFERENCES UserReactions(id) ON DELETE CASCADE
    CHECK (
        (PostId IS NOT NULL AND CommentId IS NULL AND ReactionId IS NULL) OR
        (CommentId IS NOT NULL AND PostId IS NULL AND ReactionId IS NULL) OR
        (ReactionId IS NOT NULL AND CommentId IS NULL AND PostId IS NULL)
    )
);

CREATE INDEX idx_useractivity_userid ON UserActivity(UserId);