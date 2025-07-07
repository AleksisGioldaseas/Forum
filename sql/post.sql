CREATE TABLE Post (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    UserId INTEGER,
    Title TEXT NOT NULL,
    Body TEXT NOT NULL,
    Img TEXT NOT NULL,
    likes INTEGER NOT NULL DEFAULT 0 CHECK (likes >= 0),
    dislikes INTEGER NOT NULL DEFAULT 0 CHECK (dislikes >= 0),
    RankScore INTEGER NOT NULL DEFAULT 0,
    Edited INTEGER CHECK (Edited IN (1, 0)) DEFAULT 0,

    IsSuperReport BOOLEAN NOT NULL DEFAULT 0,
    SuperReportCommentId INTEGER,
    SuperReportPostId INTEGER,
    SuperReportUserId INTEGER,

    TotalKarma INTEGER GENERATED ALWAYS AS (likes - dislikes) STORED,
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    Removed INTEGER CHECK (Removed IN (1, 0)) DEFAULT 0,
    RemovalReason TEXT,
    RemovedTime TIMESTAMP,
    ModeratorName TEXT,

    Deleted INTEGER CHECK (Deleted IN (1, 0)) DEFAULT 0,
    CommentCount INTEGER DEFAULT 0,
    -- CHECK (
    --     IsSuperReport = 0 OR (
    --         (SuperReportCommentId IS NOT NULL AND SuperReportPostId IS NULL AND SuperReportUserId IS NULL) OR
    --         (SuperReportCommentId IS NULL AND SuperReportPostId IS NOT NULL AND SuperReportUserId IS NULL) OR
    --         (SuperReportCommentId IS NULL AND SuperReportPostId IS NULL AND SuperReportUserId IS NOT NULL)
    --     )
    -- )
    FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE
);

CREATE INDEX idx_post_userid ON Post(UserId);