/*
Functions to authenticate usernames and passwords.
*/

package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
)

/*
Checks that a given username is valid.

Currently, a username is valid if it consists of letters, digits, and underscores.
*/
func validUsername(username string) bool {
	usernameLettersAndNumsOnly := regexp.MustCompile(`\W`)
	return username == usernameLettersAndNumsOnly.ReplaceAllString(username, "") && username != ""
}

/*
Checks that a given password is valid.

Currently, a password is valid if it is not the empty string (i.e. any characters are accepted).
*/
func validPassword(password string) bool {
	return password != ""
}

/*
Given an input password and a username, verify that the input password matches the user's stored password.
*/
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

/*
Convert the password to a SHA256 hash.
*/
func hashPassword(password string) []byte {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hash.Sum(nil)
}
