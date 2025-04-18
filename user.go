/*
Everything to do with users.
*/

package main

import (
	"errors"
	"fmt"
	"net"
	"slices"
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
	username      string
	passwordHash  []byte
	connections   []net.Conn
	messageGetter chan *Message
}

/*
A message is sent by a user to any number of users.
If the user wishes to send a message to all users, numRecipients should be set to 0.
*/
type Message struct {
	sender             string
	sendingConnection  net.Conn
	specificRecipients bool /* False if message is intended for all users */
	// recipients         []*User
	message string
}

func (user *User) getMessages() {
	for {
		fmt.Println(user.username + " waiting for messages")
		message := <-user.messageGetter
		fmt.Println(user.username + " recieved message: " + message.message)
		deadConnections := make([]int, 0, len(user.connections))
		for i, connection := range user.connections {
			var from string
			if user.username == message.sender {
				if message.sendingConnection == connection {
					continue
				}
				from = "(From yourself): "
			} else if message.specificRecipients {
				from = "From " + message.sender + " (direct message): "
			} else {
				from = "From " + message.sender + ": "
			}
			_, err := connection.Write([]byte(from + message.message))
			if errors.Is(err, net.ErrClosed) {
				fmt.Println("Dead connection found at " + user.username)
				deadConnections = append(deadConnections, i)
			} else if err != nil {
				fmt.Println(err.Error())
			}
		}
		user.trimDeadConnections(deadConnections)
	}
}

func (user *User) trimDeadConnections(deadConnections []int) {
	for _, i := range deadConnections {
		err := user.connections[i].Close()
		if err != nil {

		}
		user.connections = slices.Delete(user.connections, i, i+1)
	}
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
func (users *UserMap) EnrolUser(username string, connection net.Conn) (*User, bool) {
	connection.Write([]byte("Welcome, " + username + "!\n"))
	password := ""
	passwordIsValid := false
	maxAttempts := 3
	attempts := 0
	for !passwordIsValid && attempts < maxAttempts {
		connection.Write([]byte("Please create a password (max 24 chars).\n"))
		password = getPassword(connection)
		if passwordIsValid = validPassword(password); !passwordIsValid {
			connection.Write([]byte("Password cannot be blank\n"))
			fmt.Println("Invalid password creation attempt for " + username)
			attempts++
		}
	}
	if attempts >= maxAttempts {
		fmt.Println("Too many bad password attempts.")
		connection.Write([]byte("Too many bad password attempts\n"))
		return nil, false
	}
	if _, ok := users.GetUser(username); ok {
		fmt.Println(username + " enrolled from two spots simultaneously")
		connection.Write([]byte("This username has just been enrolled\n"))
		connection.Close()
		return nil, false
	}
	user := &User{
		username:      username,
		passwordHash:  hashPassword(password),
		connections:   make([]net.Conn, 0, 1),
		messageGetter: make(chan *Message, 10),
	}
	users.Lock()
	users.AddUser(user)
	users.Unlock()
	fmt.Println("New user " + username + " enrolled!")
	go user.getMessages()
	return user, true
}
