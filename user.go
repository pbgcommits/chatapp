/*
Everything to do with users.
*/

package main

import (
	"fmt"
	"net"
	"sync"
)

/*
A read-write synchronised structure, which maps usernames to Users.
*/
type UserMap struct {
	users map[string]*User
	sync.RWMutex
}

/*
A User has a username, a hashed password, and any number of active connections.
*/
type User struct {
	username     string
	passwordHash []byte
	connections  []net.Conn
}

/*
Given a username, return the corresponding user.
*/
func (users *UserMap) GetUser(user string) (*User, bool) {
	u, ok := users.users[user]
	return u, ok
}

/*
Add a new user to users.
*/
func (users *UserMap) AddUser(user *User) {
	users.users[user.username] = user
}

/*
Given a user trying to sign in from a connection,
validate their identity maxAttempts time.
Return false if they fail maxAttempts times, else true.
*/
func (user *User) SignIn(connection net.Conn, users *UserMap, maxAttempts int) bool {
	i := 0
	username := user.username
	for i < maxAttempts {
		connection.Write([]byte("Please enter your password.\n"))
		password := getPassword(connection)
		if legalSignIn := verifyPassword(username, password, users); !legalSignIn {
			fmt.Println("Failed password attempt for " + username)
			connection.Write([]byte("Incorrect password for " + username + "\n"))
			i++
		} else {
			break
		}
	}
	if i >= 3 {
		connection.Write([]byte("Too many invalid password attempts; please try again later.\n"))
		connection.Close()
		fmt.Println("Too many invalid password attempts for " + username + ": rejecting from the system")
		return false
	}
	connection.Write([]byte("Welcome back, " + username + "!\n"))
	fmt.Println("Existing user " + username + " logged in!")
	return true
}

/*
Given a username that hasn't previously joined the server,
add them to the server. This includes them creating a password.
*/
func (users *UserMap) EnrolUser(username string, connection net.Conn) (user *User) {
	connection.Write([]byte("Welcome, " + username + "!\n"))
	password := ""
	passwordIsValid := false
	for !passwordIsValid {
		connection.Write([]byte("Please create a password (max 24 chars).\n"))
		password = getPassword(connection)
		if passwordIsValid = validPassword(password); !passwordIsValid {
			connection.Write([]byte("Password cannot be blank\n"))
			fmt.Println("Invalid password creation attempt for " + username)
		}
	}
	user = &User{
		username:     username,
		passwordHash: hashPassword(password),
		connections:  make([]net.Conn, 0, 1),
	}
	users.Lock()
	users.AddUser(user)
	users.Unlock()
	fmt.Println("New user " + username + " enrolled!")
	return
}
