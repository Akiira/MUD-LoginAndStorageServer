package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	PASSWORD = 0
	ADDRESS  = 1
)

var servers map[string]string

func main() {
	servers = make(map[string]string)

	readServerList()

	runCharacterServer()
	//runClientServer()
}

func runCharacterServer() {
	listener := setUpServerWithAddress(servers["characterStorage"])

	for {
		fmt.Println("Character Server: i'm waiting")
		conn, err := listener.Accept()
		//checkError(err)
		if err == nil {
			fmt.Println("Character Server:Connection established")
			sendCharacterFile(conn)
			conn.Close()
		}
	}
}

func runClientServer() {
	listener := setUpServerWithAddress(servers["central"])

	for {
		fmt.Println("Client Server: i'm waiting")
		conn, err := listener.Accept()
		checkError(err)
		if err == nil {
			fmt.Println("Client Server:Connection established")
			go HandleLoginClient(conn)
		}
	}
}

func sendCharacterFile(conn net.Conn) {
	var msg ServerMessage
	err := gob.NewDecoder(conn).Decode(&msg)
	checkError(err)
	fmt.Println(msg)

	name := msg.Value[0].Value

	file, err := os.Open("Characters/" + name + ".xml")
	checkError(err)
	defer file.Close()

	_, err = io.Copy(conn, file)
	checkError(err)
}

func HandleLoginClient(myConn net.Conn) {

	var clientResponse ClientMessage

	err := gob.NewDecoder(myConn).Decode(&clientResponse)
	checkError(err)

	if _, err := os.Stat("Passwords/" + clientResponse.getUsername()); err == nil {
		file, err := os.Open("Passwords/" + clientResponse.getUsername())
		checkError(err)

		reader := bufio.NewReader(file)

		line, _, err := reader.ReadLine()

		s := strings.Split(string(line), " ")

		if s[PASSWORD] == clientResponse.getPassword() {
			gob.NewEncoder(myConn).Encode(newServerMessage(REDIRECT, s[ADDRESS]))
		} else {
			//TODO
			//Incorrect password
		}
	} else {
		//TODO
		//Character not found
	}

	myConn.Close()
}

func readServerList() {
	servers = make(map[string]string)
	file, err := os.Open("serverConfig/serverList.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		readData := strings.Fields(scanner.Text())
		fmt.Println(readData[0], " ", readData[1])
		servers[readData[0]] = readData[1]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func setUpServerWithAddress(addr string) *net.TCPListener {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	return listener
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
