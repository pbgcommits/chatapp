package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
)

const SERVER_ADDRESS = ":8008"

/** Connect using nc localhost 8008
 */
func main() {
	users := make(map[string]*User)
	userChan := make(chan map[string]*User, 1)
	userChan <- users
	serverActivate(userChan)
}

func serverActivate(users chan map[string]*User) {
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
		// user := make(chan string)
		go signUp(conn, users)
	}
}

func connect(user *User, userChan chan map[string]*User) {
	for {
		// t := time.Now().Add(time.Second * 5)
		// connection.SetDeadline(t)
		connection := user.connections[0]
		connection.Write([]byte("Send a message to everybody else logged in!\n"))
		b := make([]byte, 1024)
		numBytes, err := connection.Read(b)
		message := string(b[:numBytes])
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				fmt.Println("Too long since response: " + err.Error())
			} else {
				fmt.Println("Connection closed: " + err.Error())
			}
			connection.Close()
			// TODO - delete connection from users
			// users := <-userChan
			// connectionIndex := 0
			// slices.Delete(user.connections, connectionIndex, connectionIndex+1)
			// userChan <- users
			return
		}
		sendToUsers(user.username, message, userChan)
		fmt.Printf("Read from connection: %s", b)
	}
}

func signUp(newConnection net.Conn, userChan chan map[string]*User) {
	chooseUsernameMessage := "What's your username? (max 24 chars) "
	_, errW := newConnection.Write([]byte(chooseUsernameMessage))
	if errW != nil {
		fmt.Println("Err asking for username: " + errW.Error())
		return
	}
	usernameBytes := make([]byte, 24)
	numBytes, errR := newConnection.Read(usernameBytes)
	usernameBytes = bytes.TrimSpace(usernameBytes[:numBytes])
	username := string(usernameBytes)
	if errR != nil {
		fmt.Println("Error getting username: " + errR.Error())
		return
	}
	users := <-userChan
	// userChan <- users
	user, ok := users[username]
	if ok {
		// TODO: check/add password
		newConnection.Write([]byte("Welcome back, " + username + "!\n"))
	} else {
		newConnection.Write([]byte("Welcome, " + username + "!\n"))
		// users = <-userChan
		users[username] = &User{
			username:     username,
			passwordHash: "",
			connections:  make([]net.Conn, 0, 1),
		}
		user = users[username]
	}
	user.connections = append(users[username].connections, newConnection)
	fmt.Println(users[username])
	userChan <- users
	connect(users[username], userChan)
}

func sendToUsers(sender string, message string, userChan chan map[string]*User) {
	users := <-userChan
	for name, user := range users {
		if name == sender {
			continue
		}
		deadConnections := make([]int, 0, len(user.connections))
		for index, connection := range user.connections {
			_, err := connection.Write([]byte("From " + sender + ": " + message))
			if errors.Is(err, net.ErrClosed) {
				deadConnections = append(deadConnections, index)
				fmt.Println("Connection no longer exists: " + err.Error())
			} else if err != nil {
				fmt.Println("Unexpected error on write: " + err.Error())
			}
		}
		for _, deadConnection := range deadConnections {
			fmt.Printf("Deleting connection: %v\n", user.connections[deadConnection])
			fmt.Println(user.connections)
			user.connections = slices.Delete(user.connections, deadConnection, deadConnection+1)
			fmt.Println(user.connections)
		}
	}
	userChan <- users
}
