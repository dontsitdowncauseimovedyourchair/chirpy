# Chirpy API

My First API!!!!!!

Chirpy is a RESTful API written in Go. Basically Twitter 2. It supports user registration, secure authentication via JWT and refresh tokens, chirp creation and management with (not so advanced) profanity filtering, Chirpy Red subscription upgrades via webhooks, and administrative metrics.

---

## Tech Stack

- **Language:** Go (1.23+)
- **Database:** PostgreSQL
- **SQL / Migrations:** [sqlc](https://sqlc.dev/) & [Goose](https://github.com/pressly/goose)
- **Password Hashing:** Argon2id
- **Tokens:** JWT (`golang-jwt/jwt/v5`) & Secure Hex Refresh Tokens

---

## Environment Variables

To run this you'll need a `.env` file in the root directory with the following variables:

| Variable    | Description                                           | Example                                                                |
|:------------|:------------------------------------------------------|:-----------------------------------------------------------------------|
| `DB_URL`    | PostgreSQL connection string                          | `postgresql://postgres:postgres@localhost:5432/chirpy?sslmode=disable` |
| `PLATFORM`  | Environment mode (`dev` enables admin reset endpoint) | `dev`                                                                  |
| `SECRET`    | Secret key used to sign and verify JWT access tokens  | `your-jwt-secret-key`                                                  |
| `POLKA_KEY` | API key used to authorize Polka webhook events        | `your-polka-api-key`                                                   |

---

## Getting Started

### 1. Database Migrations

Run database migrations using Goose:

```bash
cd sql/schema
goose postgres "postgresql://localhost:5432/chirpy?sslmode=disable" up
cd ../..
```

*(Alternatively, run `./migrate_up.sh`)*

### 2. Run the Server

```bash
go run .
```

The server will start listening on port `8080` (default URL: `http://localhost:8080`).

---

## Authentication

The API uses three types of authentication:

1. **Access Tokens (JWT):** Short-lived tokens passed in the `Authorization` header for protected endpoints:
   ```http
   Authorization: Bearer <jwt_access_token>
   ```
2. **Refresh Tokens:** Long-lived tokens (60 days) passed in the `Authorization` header to refresh access tokens or revoke sessions:
   ```http
   Authorization: Bearer <refresh_token>
   ```
3. **Webhook API Key:** Provided in the `Authorization` header for external webhook integrations:
   ```http
   Authorization: ApiKey <POLKA_KEY>
   ```

---

## Error Handling

Failed requests return standard JSON error objects:

```json
{
  "error": "Error description message"
}
```

Common HTTP status codes:
- `400 Bad Request` – Invalid payload or malformed query/path parameter
- `401 Unauthorized` – Missing, invalid, or expired token / API key
- `403 Forbidden` – Authenticated user lacks permission for the resource
- `404 Not Found` – Resource does not exist
- `500 Internal Server Error` – Server-side error

---

## API Endpoints

### Health & Administration

#### Check Health
Returns `200 OK` if the server is healthy.

- **Method:** `GET`
- **Path:** `/api/healthz`
- **Auth:** None
- **Response:**
  - Status: `200 OK`
  - Content-Type: `text/plain`
  - Body: `OK`

---

#### Get Admin Metrics
Renders an HTML dashboard displaying total file server hits.

- **Method:** `GET`
- **Path:** `/admin/metrics`
- **Auth:** None
- **Response:**
  - Status: `200 OK`
  - Content-Type: `text/html`

---

#### Reset Database & Metrics
Resets hits count to zero and wipes all user data. **Only accessible when `PLATFORM=dev`**.

- **Method:** `POST`
- **Path:** `/admin/reset`
- **Auth:** Dev environment check (`PLATFORM=dev`)
- **Response:**
  - Status: `200 OK` (or `403 Forbidden` if not in `dev` mode)

---

### Authentication

#### Log In
Authenticates user credentials and returns user details, an access token (JWT), and a refresh token.

- **Method:** `POST`
- **Path:** `/api/login`
- **Auth:** None
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "mysecretpassword"
  }
  ```
- **Response:**
  - Status: `200 OK`
  - Body:
    ```json
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "created_at": "2026-09-16T18:00:00Z",
      "updated_at": "2026-09-16T18:00:00Z",
      "email": "user@example.com",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "a1b2c3d4e5...",
      "is_chirpy_red": false
    }
    ```

---

#### Refresh Access Token
Issues a new 1-hour JWT access token using a valid refresh token.

- **Method:** `POST`
- **Path:** `/api/refresh`
- **Headers:**
  - `Authorization: Bearer <refresh_token>`
- **Response:**
  - Status: `200 OK`
  - Body:
    ```json
    {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
    }
    ```

---

#### Revoke Refresh Token
Revokes an active refresh token, ending the session.

- **Method:** `POST`
- **Path:** `/api/revoke`
- **Headers:**
  - `Authorization: Bearer <refresh_token>`
- **Response:**
  - Status: `204 No Content`

---

### Users

#### Create User
Registers a new user account with an email and password.

- **Method:** `POST`
- **Path:** `/api/users`
- **Auth:** None
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "mysecretpassword"
  }
  ```
- **Response:**
  - Status: `201 Created`
  - Body:
    ```json
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "created_at": "2026-09-16T18:00:00Z",
      "updated_at": "2026-09-16T18:00:00Z",
      "email": "user@example.com",
      "token": "",
      "refresh_token": "",
      "is_chirpy_red": false
    }
    ```

---

#### Update User Credentials
Updates the email and/or password of the authenticated user.

- **Method:** `PUT`
- **Path:** `/api/users`
- **Headers:**
  - `Authorization: Bearer <jwt_access_token>`
- **Request Body:**
  ```json
  {
    "email": "newemail@example.com",
    "password": "newpassword123"
  }
  ```
- **Response:**
  - Status: `200 OK`
  - Body:
    ```json
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "created_at": "2026-09-16T18:00:00Z",
      "updated_at": "2026-09-16T18:05:00Z",
      "email": "newemail@example.com",
      "token": "",
      "refresh_token": "",
      "is_chirpy_red": false
    }
    ```

---

### Chirps

#### Create Chirp
Publishes a new chirp (max 140 characters). Profane words (`KERFUFFLE`, `SHARBERT`, `FORNAX`) are automatically censored to `****` (yes, those are very bad words).

- **Method:** `POST`
- **Path:** `/api/chirps`
- **Headers:**
  - `Authorization: Bearer <jwt_access_token>`
- **Request Body:**
  ```json
  {
    "body": "Hello world from Chirpy!"
  }
  ```
- **Response:**
  - Status: `201 Created`
  - Body:
    ```json
    {
      "id": "234e5678-e89b-12d3-a456-426614174000",
      "created_at": "2026-09-16T18:10:00Z",
      "updated_at": "2026-09-16T18:10:00Z",
      "body": "Hello world from Chirpy!",
      "user_id": "123e4567-e89b-12d3-a456-426614174000"
    }
    ```

---

#### Get All Chirps
Retrieves a list of chirps with optional author filtering and sorting.

- **Method:** `GET`
- **Path:** `/api/chirps`
- **Auth:** None
- **Query Parameters:**
  | Parameter | Type | Required | Description |
  | :--- | :--- | :--- | :--- |
  | `author_id` | UUID | No | Filter chirps written by a specific user |
  | `sort` | string | No | Sort direction by creation timestamp: `asc` (default) or `desc` |
- **Response:**
  - Status: `200 OK`
  - Body:
    ```json
    [
      {
        "id": "234e5678-e89b-12d3-a456-426614174000",
        "created_at": "2026-09-16T18:10:00Z",
        "updated_at": "2026-09-16T18:10:00Z",
        "body": "Hello world from Chirpy!",
        "user_id": "123e4567-e89b-12d3-a456-426614174000"
      }
    ]
    ```

---

#### Get Chirp by ID
Retrieves a single chirp by its UUID.

- **Method:** `GET`
- **Path:** `/api/chirps/{chirpID}`
- **Auth:** None
- **Response:**
  - Status: `200 OK`
  - Body:
    ```json
    {
      "id": "234e5678-e89b-12d3-a456-426614174000",
      "created_at": "2026-09-16T18:10:00Z",
      "updated_at": "2026-09-16T18:10:00Z",
      "body": "Hello world from Chirpy!",
      "user_id": "123e4567-e89b-12d3-a456-426614174000"
    }
    ```

---

#### Delete Chirp
Deletes a chirp by its UUID. **Only the chirp's original author can delete it.**

- **Method:** `DELETE`
- **Path:** `/api/chirps/{chirpID}`
- **Headers:**
  - `Authorization: Bearer <jwt_access_token>`
- **Response:**
  - Status: `204 No Content`
  - Errors:
    - `403 Forbidden` if the authenticated user is not the author (`"not your chirp my friendo!"`)
    - `404 Not Found` if the chirp does not exist

---

### Webhooks

#### Polka Webhook
Receives events from the (totally real) Polka payment provider. Upgrades a user to the incredible subscription called Chirpy Red upon receiving the `user.upgraded` event.

- **Method:** `POST`
- **Path:** `/api/polka/webhooks`
- **Headers:**
  - `Authorization: ApiKey <POLKA_KEY>`
- **Request Body:**
  ```json
  {
    "event": "user.upgraded",
    "data": {
      "user_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
  ```
- **Response:**
  - Status: `204 No Content`
  - Errors:
    - `401 Unauthorized` if `ApiKey` header is missing or does not match `POLKA_KEY`
    - `404 Not Found` if user with `user_id` does not exist
