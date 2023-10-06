package main

import (
	"fmt"
	"net"
	"os"

	"github.com/ishidawataru/sctp"
)

func handleConnection(conn *sctp.SCTPConn) {
	defer conn.Close()

	// Send a greeting to the client
	message := "Hello, what's your name?  "
	_, err := conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message to client:", err)
		return
	}

	// Receive the name from the client
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	name := string(buffer[:n])
	response := fmt.Sprintf("Nice to see you, %s", name)
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error sending response to client:", err)
		return
	}
}

func main() {
	IPAddr := sctp.SCTPAddr{IPAddrs: []net.IPAddr{{IP: net.ParseIP("0.0.0.0")}}, Port: 38412}

	listener, err := sctp.ListenSCTP("sctp", &IPAddr)
	if err != nil {
		fmt.Println("Error creating SCTP listener:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Listening for SCTP connections on %s...\n", &IPAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting SCTP connection:", err)
			continue
		}

		sctpConn := conn.(*sctp.SCTPConn)
		go handleConnection(sctpConn)
	}
}
