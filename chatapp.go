package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
)

const SERVER_ADDRESS = ":9018"

/** Connect using nc localhost 9018
 */
func main() {
	// h := sha256.New()
	// h.Write([]byte("apple"))
	// fmt.Println(h.Sum(make([]byte, 0)))
	// h = sha256.New()
	// h.Write([]byte("apple"))
	// fmt.Println(h.Sum(make([]byte, 0)))
	// return
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
	username := ""
	usernameIsValid := false
	for !usernameIsValid {
		chooseUsernameMessage := "What's your username? (max 24 chars) "
		_, errW := newConnection.Write([]byte(chooseUsernameMessage))
		if errW != nil {
			fmt.Println("Err asking for username: " + errW.Error())
			return
		}
		usernameBytes := make([]byte, 24)
		numBytes, errR := newConnection.Read(usernameBytes)
		if errR != nil {
			fmt.Println("Error getting username: " + errR.Error())
			return
		}
		usernameBytes = bytes.TrimSpace(usernameBytes[:numBytes])
		username = string(usernameBytes)
		if usernameIsValid = validUsername(username); !usernameIsValid {
			_, errW := newConnection.Write([]byte("Invalid username (only use letters, numbers, and underscores)\n"))
			if errW != nil {
				fmt.Println("Err informing username is invalid: " + errW.Error())
				return
			}
		}
	}
	users := <-userChan
	user, ok := users[username]
	userChan <- users
	if ok {
		i := 0
		for i < 3 {
			newConnection.Write([]byte("Please enter your password.\n"))
			password := getPassword(newConnection)
			if legalSignIn := verifyPassword(username, password, userChan); !legalSignIn {
				fmt.Println("Failed password attempt for " + username)
				newConnection.Write([]byte("Incorrect password for " + username + "\n"))
				i++
			} else {
				break
			}
		}
		if i >= 3 {
			newConnection.Write([]byte("Too many invalid password attempts; please try again later.\n"))
			newConnection.Close()
			fmt.Println("Too many invalid password attempts for " + username + ": rejecting from the system")
			return
		}
		newConnection.Write([]byte("Welcome back, " + username + "!\n"))
		fmt.Println("Existing user " + username + " logged in!")
	} else {
		newConnection.Write([]byte("Welcome, " + username + "!\n"))
		password := ""
		passwordIsValid := false
		for !passwordIsValid {
			newConnection.Write([]byte("Please create a password (max 24 chars).\n"))
			password = getPassword(newConnection)
			for passwordIsValid = validPassword(password); !passwordIsValid; {
				newConnection.Write([]byte("Password cannot be blank\n"))
				fmt.Println("Invalid password creation attempt for " + username)
			}
		}
		users = <-userChan
		users[username] = &User{
			username:     username,
			passwordHash: hashPassword(password),
			connections:  make([]net.Conn, 0, 1),
		}
		user = users[username]
		userChan <- users
		fmt.Println("New user " + username + " enrolled!")
	}
	users = <-userChan
	user.connections = append(users[username].connections, newConnection)
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
	fmt.Println("password: " + password)
	return password
}
