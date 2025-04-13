package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net"
	"regexp"
)

type User struct {
	username     string
	passwordHash []byte
	connections  []net.Conn
}

func validUsername(username string) bool {
	usernameLettersAndNumsOnly := regexp.MustCompile(`\W`)
	return username == usernameLettersAndNumsOnly.ReplaceAllString(username, "") && username != ""
}

func validPassword(password string) bool {
	return password != ""
}

func verifyPassword(username string, password string, userChan chan map[string]*User) bool {
	hashedInput := hashPassword(password)
	correctPassword := false
	users := <-userChan
	user, ok := users[username]
	if ok {
		fmt.Println(user.passwordHash)
		fmt.Println(hashedInput)
		correctPassword = bytes.Equal(hashedInput, user.passwordHash)
	} else {
		fmt.Println("Error - invalid username used for password check")
	}
	userChan <- users
	return correctPassword
}

func hashPassword(password string) []byte {
	hash := sha256.New()
	fmt.Println([]byte(password))
	hash.Write([]byte(password))
	return hash.Sum(nil)
}
