package auth

import (
	"github.com/google/uuid"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	signingKey := []byte(tokenSecret)

	claims := jwt.RegisteredClaims {
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "chirpy-access",
			Subject:   userID.String(),
		}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	key, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return key, nil
}
