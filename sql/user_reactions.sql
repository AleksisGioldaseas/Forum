CREATE TABLE UserReactions (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    UserId INTEGER NOT NULL,
    PostId INTEGER,
    CommentId INTEGER,
    Reaction INTEGER CHECK (Reaction IN (1, 0, -1)),
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE,
    FOREIGN KEY (PostId) REFERENCES Post(id) ON DELETE CASCADE,
    FOREIGN KEY (CommentId) REFERENCES Comment(id) ON DELETE CASCADE,
    CHECK (
        (PostId IS NOT NULL AND CommentId IS NULL) OR
        (PostId IS NULL AND CommentId IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_user_reactions_user_post_unique ON UserReactions(UserId, PostId)
WHERE PostId IS NOT NULL;

CREATE UNIQUE INDEX idx_user_reactions_user_comment_unique ON UserReactions(UserId, CommentId)
WHERE CommentId IS NOT NULL;

CREATE INDEX idx_userreactions_user ON UserReactions(UserId);
CREATE INDEX idx_userreactions_post ON UserReactions(PostId);
CREATE INDEX idx_userreactions_comment ON UserReactions(CommentId);