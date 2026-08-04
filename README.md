# Matcha

Matcha is a React and Go application backed by PostgreSQL. The backend uses handwritten SQL through Go's `database/sql` package. It does not use GORM or another ORM.

This README documents the current project, explains the database choices, and gives the commands needed to run and verify it.

## Current state

Implemented backend areas:

- registration, login, token refresh, logout, and authenticated-route verification;
- profile completion, profile/account updates, and password updates;
- profile and profile-feed reads;
- avatar upload and replacement;
- profile-photo upload and deletion with a five-photo limit;
- conversations, stored chat messages, read receipts, and WebSocket delivery;
- PostgreSQL migrations through Goose;
- Handler → Service → Repository separation for `user` and `chat`;
- local image storage behind an `ImageStorage` interface;
- unit tests for image services and local storage;
- PostgreSQL integration tests for chat and profile-photo repositories.

Not implemented yet:

- notification handlers/services/repositories. The table exists, but its routes in `backend/main.go` remain commented out;
- production-grade object storage and background cleanup of orphaned files;
- a complete automated test suite for every user/account endpoint.

## Architecture

The backend uses this direction of dependency:

```text
HTTP / WebSocket request
        ↓
Handler
        ↓
Service
        ↓
Repository or ImageStorage
        ↓
PostgreSQL or local filesystem
```

### Handler

A handler owns HTTP-specific work:

- reading authentication data from the request context;
- decoding JSON or multipart forms;
- parsing path/query values;
- calling a service;
- translating domain errors into HTTP status codes;
- encoding a response.

A handler does not contain SQL and does not write files directly.

### Service

A service owns use-case and business logic:

- normalization and validation;
- password hashing and comparison;
- enforcing profile and image rules;
- coordinating repository and file-storage operations;
- cleaning up a newly written file when a later database operation fails.

A service does not know about `http.ResponseWriter`.

### Repository

A repository owns persistence logic:

- SQL strings;
- `$1`, `$2`, … query arguments;
- `QueryRowContext`, `QueryContext`, and transactions;
- exact `Scan` order;
- conversion of `sql.ErrNoRows` and PostgreSQL constraint errors into package errors.

Only `PostgresRepository` stores `*sql.DB`. `UserHandler` does not have a database field.

### Models

`backend/models` contains shared data structures such as `User`, `Photo`, `Conversation`, and `Message`. Models do not own SQL queries, HTTP behavior, or package-specific errors.

### Package layout

The important backend layout is:

```text
backend/
├── api/                       token creation and refresh
├── chat/
│   ├── handler*.go
│   ├── service*.go
│   ├── repository*.go
│   └── *_test.go
├── database/
│   ├── migrations/            Goose migration history
│   ├── schema/                per-table schema reference
│   └── schema.sql             reference schema entry point
├── middleware/                authentication middleware
├── models/                    shared data structures
├── realtime/                  WebSocket clients and hub
├── user/
│   ├── handler*.go
│   ├── service*.go
│   ├── repository*.go
│   ├── image_storage.go
│   ├── local_image_storage.go
│   └── *_test.go
├── db.go
├── entrypoint.sh
└── main.go
```

## Running the project

### Environment

Create `.env` from `.env_example` and provide at least the variables used by the current stack:

```text
POSTGRES_USER
POSTGRES_PASSWORD
POSTGRES_DB

SQL_DATABASE
SQL_USER
SQL_PASSWORD
SQL_HOST
SQL_PORT

PGADMIN_DEFAULT_EMAIL
PGADMIN_DEFAULT_PASSWORD
SECRET_KEY
DELETE_DB
```

Inside Docker, `SQL_HOST` should be `db` and `SQL_PORT` should be `5432`. The `POSTGRES_*` values initialize the container; the matching `SQL_*` values are used by Go and Goose.

`DELETE_DB=yes` makes `entrypoint.sh` drop and recreate the development database before applying migrations. This destroys its data. Normally use:

```text
DELETE_DB=no
```

### Docker Compose

From the repository root:

```sh
docker compose up --build
```

The stack contains:

- PostgreSQL;
- pgAdmin;
- Go backend with Air live reload;
- React frontend;
- Redis (reserved for later use);
- nginx with local TLS termination.

Useful checks:

```sh
docker compose ps
docker compose logs web
docker compose logs db
```

The backend listens on port `8000` inside the stack. nginx exposes it under `/backend/` and proxies the frontend at `/`.

## Database initialization and Goose

`backend/entrypoint.sh` waits for PostgreSQL and automatically runs:

```sh
goose -dir ./database/migrations postgres "$DATABASE_DSN" up
```

The active migration history is stored in:

```text
backend/database/migrations/
```

The current initial migration is:

```text
backend/database/migrations/00001_initial_schema.sql
```

Goose is a migration tool, not an ORM. It executes SQL written in migration files and records applied versions in `goose_db_version`.

Check migration status inside the backend container:

```sh
docker compose exec web sh -lc '
goose -dir ./database/migrations postgres \
  "host=$SQL_HOST port=$SQL_PORT user=$SQL_USER password=$SQL_PASSWORD dbname=$SQL_DATABASE sslmode=disable" \
  status
'
```

Apply pending migrations manually:

```sh
docker compose exec web sh -lc '
goose -dir ./database/migrations postgres \
  "host=$SQL_HOST port=$SQL_PORT user=$SQL_USER password=$SQL_PASSWORD dbname=$SQL_DATABASE sslmode=disable" \
  up
'
```

Create a new migration:

```sh
docker compose exec web goose -dir ./database/migrations create describe_change sql
```

Once a migration has been applied to a shared or important database, do not edit that migration. Add the next numbered migration.

`backend/database/schema/` is a readable per-table reference for a fresh schema. Runtime startup uses Goose migrations, so every real schema change must be represented by a new migration. Keep the reference schema synchronized if it remains in the project.

## Current tables

### `users`

Stores account, authentication, and profile data:

- identity and timestamps;
- username, names, email, and bcrypt password hash;
- profile completion state;
- gender, preferences, biography, and JSONB interests;
- avatar URL.

### `photos`

Stores profile-photo metadata. The file itself is stored outside PostgreSQL; the table stores its URL and owner.

```text
photos.user_id → users.id
```

Deleting a user cascades to their photo rows.

### `conversations`

Stores exactly one ordered pair of users. The lower ID is stored in `user_one_id`, and the higher ID is stored in `user_two_id`.

This rule:

```sql
CHECK (user_one_id < user_two_id)
```

together with:

```sql
UNIQUE (user_one_id, user_two_id)
```

prevents both duplicate orientations `(5, 10)` and `(10, 5)`.

### `chat_messages`

Stores persisted messages, participants, timestamps, and nullable `read_at`.

### `notifications`

The schema exists, but the application layer is intentionally not implemented yet.

## Important SQL words and symbols

### `PRIMARY KEY`

Uniquely identifies a row. PostgreSQL automatically creates a unique B-tree index for a primary key.

```sql
id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY
```

### `GENERATED ... AS IDENTITY`

Asks PostgreSQL to generate the next ID when an insert omits `id`. Application inserts normally do not provide the ID themselves.

### `NOT NULL`

Forbids SQL `NULL`. It does not forbid an empty string, so important text fields also use a `CHECK` constraint.

### `DEFAULT`

Supplies a value when an insert omits the column:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

### `UNIQUE`

Forbids duplicate values. PostgreSQL creates a unique index to enforce it.

```sql
email TEXT NOT NULL UNIQUE
```

For several columns, uniqueness applies to the combination:

```sql
UNIQUE (user_one_id, user_two_id)
```

### `CONSTRAINT`

Gives a database rule a stable name:

```sql
CONSTRAINT photos_url_not_blank
CHECK (btrim(url) <> '')
```

The name does not create a new kind of rule. It names the `CHECK`, `UNIQUE`, or foreign-key rule that follows. Named constraints are useful because PostgreSQL includes the name in an error and future migrations can refer to it explicitly.

### `CHECK`

Requires an expression to be true before PostgreSQL accepts a row:

```sql
CHECK (user_one_id <> user_two_id)
```

Application validation gives a friendly response. A database `CHECK` is the final protection against bugs, concurrent requests, and manual SQL.

### `<>`, `<`, and `>`

- `<>` means “not equal” in SQL;
- `<` means “less than”;
- `>` means “greater than”.

For example:

```sql
btrim(url) <> ''
```

means that after removing surrounding spaces, the URL must not be an empty string.

Angle brackets in explanatory placeholders such as `<user-id>` mean “replace this text with a real value”. They are not part of the final command. In actual SQL, `<` and `>` are comparison operators.

### Foreign keys and `REFERENCES`

```sql
user_id BIGINT NOT NULL
    REFERENCES users(id)
    ON DELETE CASCADE
```

This has two separate jobs:

1. `REFERENCES users(id)` rejects a photo whose user does not exist.
2. `ON DELETE CASCADE` deletes that user's photo rows when the user row is deleted.

A foreign key protects relationships. It does not automatically create an index on the referencing column, so `photos.user_id` has its own index.

### `ON DELETE SET NULL`

Keeps the child row but replaces its foreign key with `NULL` when the referenced row is deleted. `notifications.sender_id` uses this because a notification can remain meaningful after its sender disappears.

### `TIMESTAMPTZ`

Stores an absolute moment in time. PostgreSQL converts display values according to the session time zone. Go scans it into `time.Time`.

### `JSONB`

Stores validated JSON in a binary PostgreSQL representation. Interests use an array:

```sql
interests JSONB NOT NULL DEFAULT '[]'::jsonb
```

The constraint confirms the value is an array:

```sql
CHECK (jsonb_typeof(interests) = 'array')
```

### `COALESCE`

Returns the first non-`NULL` value:

```sql
gender = COALESCE($1, gender)
```

If `$1` is not `NULL`, the column receives `$1`. If `$1` is `NULL`, the existing `gender` value is retained. The update services use pointer fields so omitted JSON properties become SQL `NULL` while explicitly provided strings remain distinguishable.

### `RETURNING`

Returns columns from a row changed by `INSERT`, `UPDATE`, or `DELETE`:

```sql
INSERT INTO photos (user_id, url)
VALUES ($1, $2)
RETURNING id, user_id, url
```

It is not always required. It is useful when Go immediately needs a generated ID, timestamp, updated value, or deleted file URL. Use `QueryRowContext(...).Scan(...)` when a query has `RETURNING`.

## Indexes

An index is a separate PostgreSQL data structure, not an extra visible column inside the table. PostgreSQL's default index type is a B-tree.

The table may conceptually contain:

```text
photos
id    user_id    url
1     42         /uploads/photos/a.jpg
2     17         /uploads/photos/b.jpg
3     42         /uploads/photos/c.jpg
```

The index:

```sql
CREATE INDEX photos_user_id_idx ON photos(user_id);
```

conceptually keeps searchable entries similar to:

```text
17 → photo row 2
42 → photo row 1
42 → photo row 3
```

It does not require `user_id` to be unique. Several entries can point to different photo rows with the same user ID. PostgreSQL walks the balanced B-tree to find the `42` section and then follows those entries to the matching table rows instead of scanning every photo.

Indexes cost disk space and make writes slightly more expensive because PostgreSQL must update both the table and its indexes. Add them for actual filtering, joins, ordering, and uniqueness—not for every column automatically.

### Composite indexes

```sql
CREATE INDEX chat_messages_conversation_created_idx
ON chat_messages(conversation_id, created_at);
```

This is primarily ordered by `conversation_id`, then by `created_at` inside each conversation. It supports queries such as:

```sql
WHERE conversation_id = $1
ORDER BY created_at ASC
```

Column order matters. A composite index beginning with `(user_one_id, user_two_id)` can efficiently search by `user_one_id` alone, but it normally does not replace an index whose first searchable column is `user_two_id`.

### Partial indexes

```sql
CREATE INDEX chat_messages_recipient_unread_idx
ON chat_messages(recipient_id)
WHERE read_at IS NULL;
```

Only unread messages appear in this index, so it is smaller and directly matches unread-message queries.

### `ASC` and `DESC`

- `ASC` orders from smaller/earlier to larger/later;
- `DESC` orders from larger/later to smaller/earlier.

```sql
ORDER BY created_at DESC, id DESC
```

returns newest rows first. Adding `id` gives deterministic order when two rows share the same timestamp.

## Writing queries with `database/sql`

### Placeholders

PostgreSQL uses `$1`, `$2`, and so on:

```sql
SELECT id, password, is_completed
FROM users
WHERE email = $1
```

Pass values separately. Never concatenate request data into SQL.

### One row: `QueryRowContext`

```go
err := repo.db.QueryRowContext(ctx, query, email).Scan(
    &user.ID,
    &user.Password,
    &user.IsCompleted,
)
```

`QueryRowContext` means “the caller expects one row”, while `Context` allows cancellation when the request disconnects or reaches a deadline. The number and order of `Scan` destinations must exactly match the selected columns.

No matching row is reported by `Scan` as `sql.ErrNoRows`.

### Several rows: `QueryContext`

```go
rows, err := repo.db.QueryContext(ctx, query, userID)
if err != nil {
    return nil, err
}
defer rows.Close()

for rows.Next() {
    var photo models.Photo
    if err := rows.Scan(&photo.ID, &photo.URL); err != nil {
        return nil, err
    }
    photos = append(photos, photo)
}

if err := rows.Err(); err != nil {
    return nil, err
}
```

Always close rows and check `rows.Err()` after iteration.

### `make` and `append`

```go
photos := make([]models.Photo, 0, 5)
photos = append(photos, photo)
```

This creates an empty slice with capacity reserved for five elements. Its length is still zero. `append` adds actual elements. Returning an initialized empty slice produces `[]` in JSON instead of `null`.

### Transactions and `FOR UPDATE`

`CreatePhotos` performs these operations in one transaction:

1. lock the user row with `SELECT ... FOR UPDATE`;
2. count existing photos;
3. confirm `existing + new <= 5`;
4. insert every photo row;
5. commit.

The row lock makes simultaneous photo changes for the same user wait for one another, preventing both requests from accepting stale counts.

The standard pattern is:

```go
tx, err := repo.db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// Use tx.QueryRowContext / tx.QueryContext / tx.ExecContext.

if err := tx.Commit(); err != nil {
    return err
}
```

The deferred rollback is harmless after a successful commit and protects every earlier error return.

### `ON CONFLICT`

Chat creates or retrieves one conversation for an ordered user pair:

```sql
INSERT INTO conversations (user_one_id, user_two_id)
VALUES ($1, $2)
ON CONFLICT ON CONSTRAINT conversations_user_pair_unique
DO UPDATE SET updated_at = now()
RETURNING id
```

If the pair does not exist, PostgreSQL inserts it. If the named unique constraint detects the pair, only the `DO UPDATE` branch runs. `RETURNING id` gives Go the conversation ID in either case.

## Image-storage consistency

PostgreSQL and the filesystem cannot share one real transaction. The service therefore uses compensating cleanup.

Avatar replacement:

1. validate bytes and detect JPEG/PNG/WebP;
2. read the old avatar URL;
3. save the new file;
4. update PostgreSQL;
5. if the database update fails, delete the new file;
6. if it succeeds, attempt to delete the old file.

Photo upload:

1. validate every file before saving any file;
2. save all files through `ImageStorage`;
3. insert all photo rows through one repository transaction;
4. if saving or SQL fails, delete every new file already written.

Photo deletion removes the database row first and then attempts file deletion. A file-deletion failure is logged because the API-visible photo is already deleted. A production system could retry orphan cleanup in a background job.

The current implementation writes under:

```text
backend/uploads/avatars/
backend/uploads/photos/
```

and returns URLs under:

```text
/uploads/avatars/
/uploads/photos/
```

The frontend prefixes these URLs with `/backend`, allowing nginx to proxy them to Go's file server.

## Chat and realtime delivery

Messages are saved before they are delivered through the WebSocket hub. The service/repository path provides persistence; the realtime package provides delivery to currently connected clients.

One user can have several clients (for example, two browser tabs). The hub therefore stores:

```go
map[uint]map[*Client]struct{}
```

The outer key is the user ID. The inner map is the set of that user's live connections.

WebSocket event example:

```json
{
  "type": "chat_message",
  "recipient_id": 2,
  "message": "Hello"
}
```

The server saves it, builds the outgoing event, and sends it to both sender and recipient connections.

## Active HTTP routes

Public routes:

```text
POST /api/register
POST /api/login
POST /api/accounts/token/refresh
```

Authenticated routes are mounted below `/api/accounts`:

```text
GET    /verify_login
GET    /profile
GET    /profiles/feed
GET    /ws
GET    /conversations
GET    /conversations/{conversationID}/messages

POST   /logout/
POST   /profile/complete
POST   /avatar
POST   /photos

PATCH  /profile
PATCH  /user
PATCH  /user/password
PATCH  /messages/{messageID}/read

DELETE /photos/{photoID}
```

For example, the externally proxied profile endpoint is:

```text
/backend/api/accounts/profile
```

Go `http.ServeMux` variables use `{photoID}`, not `:photoID`.

## Error handling

Package errors live in `user/errors.go` and `chat/errors.go`. Services and repositories return these stable errors; handlers map them to HTTP responses.

Typical mapping:

| Situation | HTTP status |
| --- | --- |
| Invalid JSON, ID, or field | `400 Bad Request` |
| Missing/invalid authentication | `401 Unauthorized` |
| Authenticated but forbidden | `403 Forbidden` |
| User/photo/message not found | `404 Not Found` |
| Duplicate account or photo limit | `409 Conflict` |
| Unsupported image type | `415 Unsupported Media Type` |
| Oversized upload | `413 Request Entity Too Large` |
| Unexpected database/storage error | `500 Internal Server Error` |

Do not return raw PostgreSQL or filesystem errors to the browser. Log internal detail and return a stable public message.

## Verification

### Backend formatting and static checks

From `backend/`:

```sh
go fmt ./...
GOCACHE=/tmp/matcha-go-cache go test ./...
GOCACHE=/tmp/matcha-go-cache go vet ./...
GOCACHE=/tmp/matcha-go-cache go build ./...
```

### Unit tests without PostgreSQL

Integration tests skip in short mode:

```sh
GOCACHE=/tmp/matcha-go-cache go test -short ./...
```

### PostgreSQL integration tests

The repository tests accept either package-specific DSNs:

```text
USER_TEST_DATABASE_DSN
CHAT_TEST_DATABASE_DSN
```

or the normal `SQL_HOST`, `SQL_PORT`, `SQL_USER`, `SQL_PASSWORD`, and `SQL_DATABASE` variables.

Each test creates a temporary PostgreSQL schema, applies the initial migration there, runs assertions, and drops the schema afterward. The configured database user must be allowed to create schemas.

With Docker running, the simplest command is:

```sh
docker compose exec web go test ./...
```

To run only photo repository integration tests:

```sh
docker compose exec web go test ./user -run TestPostgresPhotoRepositoryIntegration -v
```

### Frontend

From `frontend/`:

```sh
npm run lint
npm run build
```

### Docker configuration

From the repository root:

```sh
docker compose config
docker compose build
docker compose up
```

### Manual API checklist

After automated checks pass, verify behavior through the running application:

- registration succeeds and duplicate username/email returns `409`;
- login returns the same `401` message for unknown email and wrong password;
- profile completion and partial updates round-trip interests correctly;
- an empty photo list is JSON `[]`, not `null`;
- invalid image bytes are rejected even if the filename looks valid;
- no avatar/photo exceeds 5 MiB;
- a profile never exceeds five photos;
- one user cannot delete another user's photo;
- only conversation participants can list its messages;
- only the recipient can mark a message as read;
- WebSocket messages are stored before realtime delivery.

Useful PostgreSQL inspection queries:

```sql
SELECT id, user_name, email, is_completed
FROM users
ORDER BY id;

SELECT id, user_id, url
FROM photos
ORDER BY user_id, id;

SELECT id, user_one_id, user_two_id
FROM conversations
ORDER BY id;

SELECT id, conversation_id, sender_id, recipient_id, read_at
FROM chat_messages
ORDER BY id;
```

## Troubleshooting

### `sql: unknown driver "pgx"`

`backend/db.go` must import the pgx standard-library adapter for its registration side effect:

```go
import _ "github.com/jackc/pgx/v5/stdlib"
```

The blank identifier is intentional. The package registers the `pgx` driver used by:

```go
sql.Open("pgx", dsn)
```

### Goose says a table already exists

Do not combine manual schema application and Goose migration application against the same fresh database without understanding their version state. Runtime initialization uses Goose. If the database is disposable, recreate it deliberately and let Goose apply the migration once.

### Docker network pool overlaps another network

The compose file lets Docker allocate network subnets automatically. Remove old unused networks if Docker still reports an overlap:

```sh
docker network prune
```

Review the prompt carefully because this removes unused Docker networks globally.

### Uploaded image returns `404`

Confirm all three pieces agree:

1. `LocalImageStorage` writes below `./uploads`;
2. `main.go` serves `/uploads/` from that directory;
3. the browser requests the proxied URL with `/backend` in front.

## Mental model

For every new database feature, write down:

1. which table owns the data;
2. which columns the query must read or change;
3. which conditions belong in `WHERE`;
4. whether one row or several rows are expected;
5. the exact order of `Scan` destinations;
6. whether concurrent operations require a transaction or row lock;
7. which layer owns validation, persistence, and HTTP mapping.

That is the practical replacement for relying on an ORM to make those decisions implicitly.
