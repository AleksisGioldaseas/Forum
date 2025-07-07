CREATE TABLE ModeratorLog (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    ActionType TEXT NOT NULL CHECK(ActionType IN ('approve', 'remove')),
    ModeratorId INTEGER NOT NULL,
    TableName TEXT NOT NULL,
    RowId INTEGER NOT NULL,
    Body TEXT,
    Created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ModeratorId) REFERENCES User(Id) ON DELETE CASCADE
);

CREATE INDEX idx_moderator_id ON ModeratorLog(ModeratorId);