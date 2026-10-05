package auth

import (
	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {

	//hash password
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	// check for errors
	if err != nil {
		return "", err
	}
	//return hashed password
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	
	// compare the password
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	
	// check for errors
	if err != nil {
		return false, err
	}
	
	// check if password is matching hash
	if !match {
		return false, nil
	}
	return match, nil
}

