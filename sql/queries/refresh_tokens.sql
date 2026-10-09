-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
  token,
  created_at,
  updated_at,
  user_id,
  expires_at,
  revoked_at
)
  VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
  );

-- name: GetRefreshToken :one
SELECT * from refresh_tokens
WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $1
WHERE token = $2;
