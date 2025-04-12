package main

import (
	"net"
)

type User struct {
	username     string
	passwordHash string
	connections  []net.Conn
}
