package main

import (
	"net"
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
	usernameRequest := make([]byte, 0)
	_, err := connection.Read(usernameRequest)
	if err != nil {
		t.Errorf("Failed to get username request: %v", err.Error())
	}
	_, err = connection.Write([]byte(username))
	if err != nil {
		t.Errorf("Failed to send username request: %v", err.Error())
	}
	return username
}

func choosePassword(t *testing.T, connection net.Conn, password string) string {
	passwordRequest := make([]byte, 0)
	for i := range 2 {
		i += 0 // idk why it complains about the underscore
		_, err := connection.Read(passwordRequest)
		if err != nil {
			t.Errorf("Failed to get password request: %v", err.Error())
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
	connection := serverConnect(t)
	connection.Close()
}

func TestUsername(t *testing.T) {
	activateServer()
	connection := serverConnect(t)
	chooseUsername(t, connection, "1")
	connection.Close()
}

func TestPassword(t *testing.T) {
	activateServer()
	connection1 := serverConnect(t)
	username1 := chooseUsername(t, connection1, "1")
	choosePassword(t, connection1, username1)
	connection2 := serverConnect(t)
	username2 := chooseUsername(t, connection2, "2")
	choosePassword(t, connection2, username2)
	connection1.Close()
	connection2.Close()
}
