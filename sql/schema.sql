PRAGMA foreign_keys = ON;

PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 10000;

.read sql/users.sql
.read sql/post.sql
.read sql/comment.sql
.read sql/categories.sql
.read sql/images.sql
.read sql/moderator.sql
.read sql/modrequests.sql
.read sql/user_reactions.sql
.read sql/notifications.sql
.read sql/removed_posts.sql
.read sql/reports.sql
.read sql/user_sessions.sql
.read sql/user_activity.sql
.read sql/triggers.sql