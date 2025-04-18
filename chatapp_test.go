package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

var serverOn bool = false

func activateServer() {
	if !serverOn {
		go main()
		serverOn = true
	}
}

func serverConnect(t *testing.T) (connection net.Conn) {
	time.Sleep(time.Second)
	connection, err := net.Dial("tcp", ":9018")
	if err != nil {
		t.Errorf("Connection failed: %v", err.Error())
	}
	return connection
}

func chooseUsername(t *testing.T, connection net.Conn, username string) string {
	usernameRequest := make([]byte, 1024)
	_, err := connection.Read(usernameRequest)
	if err != nil {
		t.Errorf("Failed to get username request: %v", err.Error())
	}
	fmt.Printf("From server: %s", usernameRequest)
	_, err = connection.Write([]byte(username))
	if err != nil {
		t.Errorf("Failed to send username request: %v", err.Error())
	}
	return username
}

func getMessage(username string, connection net.Conn, buf []byte, t *testing.T) string {
	nB, err := connection.Read(buf)
	if err != nil {
		t.Errorf("Read error: %v", err.Error())
	}
	message := string(buf[:nB])
	fmt.Println(username + " received the following: " + message)
	return message
}

func choosePassword(t *testing.T, connection net.Conn, password string) string {
	// The server sometimes sends its password request in a single message
	// and sometimes in multiple, hence the strange check for the last message
	for {
		passwordRequest := make([]byte, 1024)
		_, err := connection.Read(passwordRequest)
		if err != nil {
			t.Errorf("Failed to get password request: %v", err.Error())
		}
		splitBySpaces := strings.Split(string(passwordRequest), " ")
		// -2, as the last word is apparently a bunch of empty space
		fmt.Printf("From server: %s\n", passwordRequest)
		if splitBySpaces[len(splitBySpaces)-2] == "password:" {
			break
		}
	}
	_, err := connection.Write([]byte(password))
	if err != nil {
		t.Errorf("Failed to send password request: %v", err.Error())
	}
	return password
}

// func TestServerConnection(t *testing.T) {
// 	activateServer()
// 	_ = serverConnect(t)
// }

// func TestUsername(t *testing.T) {
// 	activateServer()
// 	connection := serverConnect(t)
// 	chooseUsername(t, connection, "1")
// }

// func TestPassword(t *testing.T) {
// 	activateServer()
// 	connection1 := serverConnect(t)
// 	username1 := chooseUsername(t, connection1, "2")
// 	choosePassword(t, connection1, username1)
// 	connection2 := serverConnect(t)
// 	username2 := chooseUsername(t, connection2, "3")
// 	choosePassword(t, connection2, username2)
// }

func TestSendToAll(t *testing.T) {
	activateServer()
	b1 := make([]byte, 1024)
	b2 := make([]byte, 1024)
	b3 := make([]byte, 1024)
	connection1 := serverConnect(t)
	username1 := chooseUsername(t, connection1, "4")
	choosePassword(t, connection1, username1)
	getMessage(username1, connection1, b1, t)
	connection2 := serverConnect(t)
	username2 := chooseUsername(t, connection2, "5")
	choosePassword(t, connection2, username2)
	getMessage(username2, connection2, b2, t)
	connection3 := serverConnect(t)
	username3 := chooseUsername(t, connection3, "4")
	choosePassword(t, connection3, username3)
	getMessage(username3, connection3, b3, t)
	connection1.Write([]byte("Hi everybody!"))
	message2 := getMessage(username2, connection2, b2, t)
	message3 := getMessage(username3, connection3, b3, t)
	/* TODO issue with testing - sometimes multiple writes are read simultaneously,
	   meaning the testing breaks :) */
	// s3 := strings.Split(message3, " ")
	// fmt.Printf("s3: %v, len: %v\n", s3, len(s3))
	// if s3[len(s3)-1] == "-h" {
	// 	fmt.Println("welcome and help sent as two messages")
	// 	message3 = getMessage(username3, connection3, b3, t)
	// }
	n2 := message2 != "From 4: Hi everybody!"
	n3 := message3 != "(From yourself): Hi everybody!"
	if n2 || n3 {
		fmt.Printf("%v, %v", n2, n3)
		t.Errorf("Send to all failed: %v, %v", message2, message3)
	}
}
