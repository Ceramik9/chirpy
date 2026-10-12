Chirpy is a Twitter like http server. This is a learning project from boot.dev.
### Installation
You will need Go programming language installed and Postgresql database. See a list of available sql migrations in sql/schema directory.

Inside a Go module:
```bash
go get github.com/ceramik9/chirpy
```

The server also requires  ".env" file in project root with below information:
```
DB_URL="database URL"
PLATFORM="dev"
SECRET="SHA-256 hash"
POLKA_KEY="32-character hexadecimal value"
```
Fill the placeholder data above.
### List of available client endpoints:

1. **GET /api/healthz** - server status

2. **POST /api/users** - create user. Request body:
```json
{
"email": "example@mail.com",
"password": "password123"
}
```
Server response:
```json
{
"id": "user id",
"created_at": timestamp,
"updated_at": timestamp,
"email": "example@mail.com",
"is_chirpy_red": false
}
```

3. **POST /api/login** - login user. Request body:
```json
{
"password": "password123",
"email": "example@mail.com"
}
```
Server response:
```json
{
"id": "user id",
"created_at": timestamp,
"updated_at": timestamp,
"email": "example@mail.com",
"token": "access token",
"refresh_token": "refresh token",
"is_chirpy_red": false
}
```

4. **PUT /api/users** - update user's email and password. Request body:
```json
{
"email": "example@mail.com",
"password": "password123"
}
```
Headers:
```
"Authorization", "Bearer {token}"
```
Server response:
```json
{
"id": "user id",
"created_at": timestamp,
"updated_at": timestamp,
"email": "example@mail.com",
"is_chirpy_red": false
}
```

5. **POST /api/chirps** - create chirp . Request body:
```json
{
"body": "text" `max characters 140`
}
```
Headers:
```
"Authorization", "Bearer {token}"
```
Server response:
```json
{
"id": "chirp id",
"created_at": timestamp,
"updated_at": timestamp,
"body": "text",
"user_id": "user id"
}
```

6. **DELETE /api/chirps/{id}** - delete chirp.
Headers:
```
"Authorization", "Bearer {token}"
```

7. **GET /api/chirps** - get all chirps. 
Available queries:
```
author_id={user id}
sort=desc/asc (sorts by created_at)
```
Server response:
```json
{[
{
"id": "chirp id",
"created_at": timestamp,
"updated_at": timestamp,
"body": "text",
"user_id": "user id"
},
0{
"id": "chirp id",
"created_at": timestamp,
"updated_at": timestamp,
"body": "text",
"user_id": "user id"
}
]}
```

8. **GET /api/chirps/{id}** - get chirp by id.
Server response:
```json
{
"id": "chirp id",
"created_at": timestamp,
"updated_at": timestamp,
"body": "text",
"user_id": "user id"
}
```

9. **POST /api/refresh** - refresh access token.
Headers:
```
"Authorization", "Bearer {refresh token}"
```
Server response:
```json
{
"token": "access token"
}
```

10. **POST /api/revoke** - revoke refresh token.
Headers:
```
"Authorization", "Bearer {refresh token}"
```


