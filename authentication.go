package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
)

func validUsername(username string) bool {
	usernameLettersAndNumsOnly := regexp.MustCompile(`\W`)
	return username == usernameLettersAndNumsOnly.ReplaceAllString(username, "") && username != ""
}

func validPassword(password string) bool {
	return password != ""
}

func verifyPassword(username string, password string, users *UserMap) bool {
	hashedInput := hashPassword(password)
	correctPassword := false
	users.RLock()
	user, ok := users.GetUser(username)
	users.RUnlock()
	if ok {
		correctPassword = bytes.Equal(hashedInput, user.passwordHash)
	} else {
		fmt.Println("Error - invalid username used for password check")
	}
	return correctPassword
}

func hashPassword(password string) []byte {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hash.Sum(nil)
}
