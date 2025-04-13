/*
Start and run a chat app server.
Currently hardcoded to run on localhost:9018.
*/

package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
)

const SERVER_ADDRESS = ":9018"
const HELP_MESSAGE = `Send a message to all users by default
Send messages to specific users by typing -u user1:user2:(etc) (message)
Get this help message by typing -h
`

/** Connect using nc localhost 9018
 */
func main() {
	users := &UserMap{
		users: make(map[string]*User),
	}
	serverActivate(users)
}

/*
Start up the server.
*/
func serverActivate(users *UserMap) {
	listener, err := net.Listen("tcp", SERVER_ADDRESS)
	if err != nil {
		fmt.Println("Error initialising server: " + err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Server initialised")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("New connection made")
		go connectToServer(conn, users)
	}
}

/*
Given a newConnection, let them either log back in (if an existing user) or
sign up to the server.
*/
func connectToServer(newConnection net.Conn, users *UserMap) {
	username, err := getUsername(newConnection)
	if err != nil {
		return
	}
	users.RLock()
	user, ok := users.GetUser(username)
	users.RUnlock()
	if ok {
		if !user.SignIn(newConnection, users, 3) {
			return
		}
	} else {
		user = users.EnrolUser(username, newConnection)
	}
	users.Lock()
	user.connections = append(user.connections, newConnection)
	users.Unlock()
	connect(user, newConnection, users)
}

/*
Maintain a connection with a user.
This function is responsible for getting messages from a user, and passing it
to other functions to send off.
*/
func connect(user *User, connection net.Conn, users *UserMap) {
	connection.Write([]byte(HELP_MESSAGE))
	for {
		b := make([]byte, 1024)
		numBytes, err := connection.Read(b)
		message := string(b[:numBytes])
		if err != nil {
			fmt.Println("Connection closed: " + err.Error())
			connection.Close()
			users.Lock()
			connectionIndex := -1
			for i, conn := range user.connections {
				if conn == connection {
					connectionIndex = i
				}
			}
			user.connections = slices.Delete(user.connections, connectionIndex, connectionIndex+1)
			users.Unlock()
			return
		}
		if splitMessage := strings.Split(message, " "); len(splitMessage[0]) > 1 && splitMessage[0][:2] == "-u" {
			if len(splitMessage) < 3 {
				connection.Write([]byte(HELP_MESSAGE))
			} else {
				recipients := strings.Split(splitMessage[1], ":")
				sendToUsers(user.username, connection, strings.Join(splitMessage[2:], " "), recipients, users)
			}
		} else if splitMessage[0] == "-h\n" {
			connection.Write([]byte(HELP_MESSAGE))
			fmt.Println("sending help message")
		} else {
			sendToAllUsers(user.username, connection, message, users)
		}
		fmt.Printf("Read from user %s: %s", user.username, b)
	}
}

/*
Given a message sent from sendingConnection by sender, send it to the user with username recipientUsername.
*/
func sendToUser(sender string, sendingConnection net.Conn, message string, recipientUsername string, recipientUser *User, users *UserMap) {
	deadConnections := make([]int, 0, len(recipientUser.connections))
	for index, connection := range recipientUser.connections {
		var err error
		if connection == sendingConnection {
			continue
		}
		if recipientUsername == sender {
			_, err = connection.Write([]byte("(From yourself): " + message))
		} else {
			_, err = connection.Write([]byte("From " + sender + ": " + message))
		}
		if errors.Is(err, net.ErrClosed) {
			deadConnections = append(deadConnections, index)
			fmt.Println("Connection no longer exists: " + err.Error())
		} else if err != nil {
			fmt.Println("Unexpected error on write: " + err.Error())
		}
	}
	users.RUnlock()
	users.Lock()
	for _, deadConnection := range deadConnections {
		fmt.Printf("Deleting connection: %v\n", recipientUser.connections[deadConnection])
		fmt.Println(recipientUser.connections)
		recipientUser.connections = slices.Delete(recipientUser.connections, deadConnection, deadConnection+1)
		fmt.Println(recipientUser.connections)
	}
	users.Unlock()
	users.RLock()
}

/*
Given a message sent from sendingConnection by sender, send it to the usernames in listOfUsers.
*/
func sendToUsers(sender string, sendingConnection net.Conn, message string, listOfUsers []string, users *UserMap) {
	users.RLock()
	for _, name := range listOfUsers {
		user, _ := users.GetUser(name)
		sendToUser(sender, sendingConnection, message, name, user, users)
	}
	users.RUnlock()
}

/* Send a message to all users logged in. */
func sendToAllUsers(sender string, sendingConnection net.Conn, message string, users *UserMap) {
	users.RLock()
	for name, user := range users.users {
		sendToUser(sender, sendingConnection, message, name, user, users)
	}
	users.RUnlock()
}

/*
Get a password from connection.
*/
func getPassword(connection net.Conn) string {
	passwordBytes := make([]byte, 24)
	_, errW := connection.Write([]byte("Enter your password: "))
	if errW != nil {
		fmt.Println("Connection closed while reading username")
		return ""
	}
	numBytes, errR := connection.Read(passwordBytes)
	if errR != nil {
		fmt.Println("Connection closed while reading password")
		return ""
	}
	password := string(bytes.TrimSpace(passwordBytes[:numBytes]))
	// fmt.Println("password: " + password)
	return password
}

/*
Get a username from connection.
*/
func getUsername(connection net.Conn) (string, error) {
	username := ""
	usernameIsValid := false
	for !usernameIsValid {
		chooseUsernameMessage := "What's your username? (max 24 chars) "
		_, errW := connection.Write([]byte(chooseUsernameMessage))
		if errW != nil {
			fmt.Println("Err asking for username: " + errW.Error())
			return username, errW
		}
		usernameBytes := make([]byte, 24)
		numBytes, errR := connection.Read(usernameBytes)
		if errR != nil {
			fmt.Println("Error getting username: " + errR.Error())
			return username, errR
		}
		usernameBytes = bytes.TrimSpace(usernameBytes[:numBytes])
		username = string(usernameBytes)
		if usernameIsValid = validUsername(username); !usernameIsValid {
			_, errW := connection.Write([]byte("Invalid username (only use letters, numbers, and underscores)\n"))
			if errW != nil {
				fmt.Println("Err informing username is invalid: " + errW.Error())
				return username, errW
			}
		}
	}
	return username, nil
}
