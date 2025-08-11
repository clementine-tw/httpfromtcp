package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("read error: %v\n", err)
		}
		_, err = conn.Write([]byte(str))
		if err != nil {
			fmt.Printf("send udp error: %v\n", err)
		}
	}
}
