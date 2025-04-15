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

func TestServerConnection(t *testing.T) {
	activateServer()
	_ = serverConnect(t)
}

func TestUsername(t *testing.T) {
	activateServer()
	connection := serverConnect(t)
	chooseUsername(t, connection, "1")
}

func TestPassword(t *testing.T) {
	activateServer()
	connection1 := serverConnect(t)
	username1 := chooseUsername(t, connection1, "2")
	choosePassword(t, connection1, username1)
	connection2 := serverConnect(t)
	username2 := chooseUsername(t, connection2, "3")
	choosePassword(t, connection2, username2)
}
