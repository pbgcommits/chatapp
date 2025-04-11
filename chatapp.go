package main

import (
	"fmt"
	"net"
)

/** Connect using nc localhost 8008
 */
func main() {
	fmt.Println("Hello, world!")
	server := make(chan net.Conn)
	go server_activate(server)
	// Right now connection is only allowed to send one message - then it disconnects
	// Presumably because connection gets overwritten after the loop finishes
	// So needs to be its own goroutine of course :)
	for {
		connection := <-server
		b := make([]byte, 8)
		connection.Read(b)
		fmt.Printf("%s\n", b)
	}
}

func server_activate(server chan net.Conn) {
	listener, err := net.Listen("tcp", ":8008")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("New connection made")
		// user := make(chan string)
		server <- conn
	}
}
