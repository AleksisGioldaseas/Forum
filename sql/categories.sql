CREATE TABLE Category (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Name TEXT UNIQUE COLLATE NOCASE NOT NULL
);

CREATE INDEX idx_category_name ON Category(Name);

CREATE TABLE PostCategory (
    PostId INTEGER,
    CategoryId INTEGER,
    PRIMARY KEY (PostId, CategoryId),
    FOREIGN KEY (PostId) REFERENCES Post(Id) ON DELETE CASCADE,
    FOREIGN KEY (CategoryId) REFERENCES Category(Id) ON DELETE CASCADE
);

CREATE INDEX idx_post_category_postid ON PostCategory(PostId);
CREATE INDEX idx_post_category_categoryid ON PostCategory(CategoryId);