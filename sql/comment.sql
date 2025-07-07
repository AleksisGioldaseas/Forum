CREATE TABLE Comment (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    UserId INTEGER,
    PostId INTEGER,
    likes INTEGER NOT NULL DEFAULT 0 CHECK (likes >= 0),
    dislikes INTEGER NOT NULL DEFAULT 0 CHECK (dislikes >= 0),
    TotalKarma GENERATED ALWAYS AS (likes - dislikes) STORED,
    Body TEXT NOT NULL,
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    Edited INTEGER CHECK (Edited IN (1, 0)) DEFAULT 0,
    Removed INTEGER CHECK (Removed IN (1, 0)) DEFAULT 0,
    RemovalReason TEXT,
    RemovedTime TIMESTAMP,
    ModeratorName TEXT,
    Deleted INTEGER CHECK (Deleted IN (1, 0)) DEFAULT 0,
    FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE,
    FOREIGN KEY (PostId) REFERENCES Post(Id) ON DELETE CASCADE
);

CREATE INDEX idx_comment_post ON Comment(PostId);