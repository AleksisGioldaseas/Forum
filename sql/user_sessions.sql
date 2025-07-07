CREATE TABLE UserSessions (
    SessionToken TEXT PRIMARY KEY,
    UserID INTEGER NOT NULL,
    ExpiresAt DATETIME NOT NULL,
    FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE
);

CREATE INDEX idx_usersessions_userid ON UserSessions(UserId);
CREATE INDEX idx_usersessions_expiresat ON UserSessions(ExpiresAt);