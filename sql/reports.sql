CREATE TABLE Reports (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    SenderId INTEGER NOT NULL,
    PostId INTEGER,
    CommentId INTEGER,
    Message TEXT,
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (SenderId) REFERENCES User(Id) ON DELETE CASCADE,
    FOREIGN KEY (PostId) REFERENCES Post(Id) ON DELETE CASCADE,
    FOREIGN KEY (CommentId) REFERENCES Comment(Id) ON DELETE CASCADE,
    CHECK (
        (PostId IS NOT NULL AND CommentId IS NULL) OR
        (PostId IS NULL AND CommentId IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_reports_post_unique ON Reports(SenderId, PostId)
WHERE PostId IS NOT NULL;

CREATE UNIQUE INDEX idx_reports_comment_unique ON Reports(SenderId, CommentId)
WHERE CommentId IS NOT NULL;

-- TODO: Check queries for single index use and get rid of these if needed
CREATE INDEX idx_reports_post_id ON Reports(PostId);
CREATE INDEX idx_reports_comment_id ON Reports(CommentId);