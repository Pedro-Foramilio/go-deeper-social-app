**Project**: Go Deeper — API

- **Description**: This repository contains a Go HTTP API implementing a small social network. It was developed with the sole purpose of learning Go and building common web backend features (authentication, posts, comments, followers, email invitations, pagination, and more).

**Base URL**: All endpoints are mounted under `/v1`.

**Authentication**:
- Health check: Basic Auth (configured via environment variables).
- Most endpoints: Bearer token (JWT) authentication via the token middleware.

**Endpoints (from `cmd/api/api.go`)**

- GET `/v1/health`
	- Auth: Basic Auth
	- Description: Returns service status, environment and version.
	- Response: 200 JSON { status, env, version }

- Posts
	- POST `/v1/posts/`
		- Auth: JWT
		- Payload: `CreatePostPayload` {
			- `title` (string, required, max 255)
			- `content` (string, required, max 1000)
			- `tags` ([]string)
		}
		- Response: 201 JSON envelope with created `post` object

	- GET `/v1/posts/{postID}/`
		- Auth: JWT
		- Description: Returns a post (populates comments)
		- Response: 200 JSON envelope with `post` and its comments

	- PATCH `/v1/posts/{postID}/`
		- Auth: JWT + ownership/role check (`moderator` role required by middleware)
		- Payload: `UpdatePostPayload` {
			- `title` (optional string)
			- `content` (optional string)
			- `tags` (optional []string)
		}
		- Response: 200 JSON envelope with updated `post`

	- DELETE `/v1/posts/{postID}/`
		- Auth: JWT + ownership/role check (`admin` role required by middleware)
		- Response: 204 No Content

- Comments
	- POST `/v1/comments/`
		- Auth: JWT
		- Payload: `CreateCommentPayload` {
			- `post_id` (int64, required)
			- `content` (string, required, max 500)
		}
		- Response: 201 JSON envelope with created `comment`

- Users
	- PUT `/v1/users/activate/{token}`
		- Auth: none
		- Description: Activates a user account using the activation token (token is hashed server-side before lookup).
		- Response: 200 JSON (empty data)

	- GET `/v1/users/{userID}/`
		- Auth: JWT
		- Description: Returns user profile (user is loaded via `userContextMiddleware`).
		- Response: 200 JSON envelope with `user`

	- PUT `/v1/users/{userID}/follow`
		- Auth: JWT
		- Description: Authenticated user follows the specified user.
		- Response: 204 or 200 (handler does not return body on success)

	- PUT `/v1/users/{userID}/unfollow`
		- Auth: JWT
		- Description: Authenticated user unfollows the specified user.
		- Response: 204 or 200 (handler does not return body on success)

	- GET `/v1/users/feed`
		- Auth: JWT
		- Description: Returns the authenticated user's feed (paginated query parameters supported).
		- Query/pagination: `limit`, `offset`, `sort` (default limit=20, offset=0, sort=desc)
		- Response: 200 JSON envelope with feed items

- Authentication
	- POST `/v1/authentication/user`
		- Auth: none
		- Description: Register a new user. Creates user record, stores a hashed activation token, and sends an invitation email with a raw activation token to the user's email.
		- Payload: `RegisterUserPayload` {
			- `username` (string, required, max 100)
			- `email` (string, required, email)
			- `password` (string, required, min 3, max 72)
		}
		- Response: 200 JSON envelope with `user` and the raw `token` (used to activate)

	- POST `/v1/authentication/token`
		- Auth: none
		- Description: Authenticate user credentials and return a JWT token.
		- Payload: `CreateUserTokenPayload` {
			- `email` (string, required)
			- `password` (string, required)
		}
		- Response: 201 JSON with the JWT token

**Notes from `cmd/api` (capabilities & behavior)**

- Input/Output helpers: `json.go` centralizes JSON reads/writes, sets a 1MB request limit, disallows unknown JSON fields, and uses a standard envelope `{ "data": ... }` for successful responses and `{ "error": ... }` for errors.
- Validation: uses `go-playground/validator` with struct tags; request payloads are validated before processing.
- Error handling: `errors.go` provides helpers to write consistent JSON error responses and logging.
- Authentication flows:
	- Registration: creates user and sends an invitation email via configured mail client (Mailtrap or SendGrid implementations are in `internal/mailer`). The activation token sent in email is the raw token; the server stores only a SHA-256 hash of the token.
	- Login: validates credentials and issues JWT tokens via the `auth` package.
- Posts & comments: Basic CRUD for posts (create, read, update, delete) with ownership and role checks, and comments creation linked to posts.
- Followers: follow/unfollow functionality via a `Followers` store.
- Feed: paginated user feed is available and uses a `PaginatedFeedQuery` parsed from query parameters.
- Context middlewares: `userContextMiddleware` and `postsContextMiddleware` load entities by path params and inject them into the request context for handlers.
- Configuration & wiring (`main.go`): the app is configurable via environment variables (`ADDR`, `DB_ADDR`, `JWT_SECRET`, `FRONTEND_URL`, email/API keys, basic auth user/pass). The server uses `zap` for logging.

**Environment / runtime notes**
- Database: PostgreSQL (configured via `DB_ADDR`), connection pooling settings available in env vars.
- Mailer: Mailtrap is used by default in the code; SendGrid support is present but commented out in `main.go`.
- JWT: configured with `JWT_SECRET`, issuer and expiry in `main.go`.

**Cache / Redis**
- Optional Redis-based cache is supported and controlled by environment variables in `main.go`:
	- `REDIS_ENABLED` (bool) — enable/disable cache
	- `REDIS_ADDR` — Redis address (default: `localhost:6379`)
	- `REDIS_PASSWORD` — Redis password
	- `REDIS_DB` — Redis database number
- Wiring: when enabled, the app initializes a Redis client (`cache.NewRedisClient`) and constructs a cache storage (`cache.NewRedisStorage`) which is injected into the application (`app.cacheStorage`).
- Implementation: cache code lives under `internal/store/cache`.
	- `Storage` exposes a `Users` store with `Get(ctx, id) (*store.User, error)` and `Set(ctx, *store.User) error`.
	- `UserStore` uses `redis.Client` and stores JSON-encoded `store.User` values under keys of the form `user-%v` with a TTL of 1 minute (uses `SETEX`).
- Usage: the `getUser` helper in `cmd/api/middleware.go` is cache-aware:
	- If `REDIS_ENABLED` is false the app directly fetches users from the database (`store.Users.GetByID`).
	- If enabled, it first attempts `cacheStorage.Users.Get(ctx, userID)`.
	- On cache miss or error it fetches the user from the DB and then calls `cacheStorage.Users.Set(ctx, user)` to populate the cache.
	- Cache hits and cache-set events are logged (`cache hit for user`, `cache set for user`).
- Behavior notes:
	- The cache is used only for reading user records by ID in the token authentication flow (`AuthTokenMiddleware` -> `getUser`).
	- Cache entries have a short TTL (1 minute), so data may be briefly stale; there is no explicit invalidation logic in the middleware shown.
	- Errors while setting or reading the cache fall back to DB reads and are logged but do not block authentication.
