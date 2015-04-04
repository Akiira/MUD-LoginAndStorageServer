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

	go runCharacterServer()
	runClientServer()
}

func runCharacterServer() {
	listener := setUpServerWithAddress(servers["characterStorage"])
	fmt.Println("\tCharacter Server: i'm waiting")

	for {
		conn, err := listener.Accept()
		checkError(err)
		if err == nil {
			fmt.Println("\tCharacter Server:Connection established")
			var msg ServerMessage
			err := gob.NewDecoder(conn).Decode(&msg)
			checkError(err)
			fmt.Println(msg)

			if msg.MsgType == GETFILE {
				sendCharacterFile(conn, msg.getMessage())
			} else {
				saveCharacterFile(conn, msg.getMessage())
			}

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

func sendCharacterFile(conn net.Conn, name string) {

	file, err := os.Open("Characters/" + name + ".xml")
	checkError(err)
	defer file.Close()

	_, err = io.Copy(conn, file)
	checkError(err)
}

func saveCharacterFile(conn net.Conn, name string) {

	file, err := os.Open("Characters/" + name + ".xml")
	checkError(err)
	defer file.Close()

	_, err = io.Copy(file, conn)
	checkError(err)
}

func HandleLoginClient(myConn net.Conn) {

	var clientResponse ClientMessage

	err := gob.NewDecoder(myConn).Decode(&clientResponse)
	checkError(err)

	if _, err := os.Stat("Characters/Passwords/" + clientResponse.getUsername() + ".txt"); err == nil {
		file, err := os.Open("Characters/Passwords/" + clientResponse.getUsername() + ".txt")
		checkError(err)

		reader := bufio.NewReader(file)

		line, _, err := reader.ReadLine()

		s := strings.Split(string(line), " ")

		if s[PASSWORD] == clientResponse.getPassword() {
			newAddress := servers[s[ADDRESS]]
			gob.NewEncoder(myConn).Encode(newServerMessage(REDIRECT, newAddress))
		} else {
			//TODO
			//Incorrect password
			fmt.Println("\tERROR: incorrect password")
		}
	} else {
		//TODO
		//Character not found
		fmt.Println("\tERROR: character not found")
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
