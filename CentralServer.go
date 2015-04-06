package main

import (
	"bufio"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"io/ioutil"
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
	//	go runCharacterServer()
	//	runClientServer()
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
			gobDecoder := gob.NewDecoder(conn)
			gobEncoder := gob.NewEncoder(conn)
			err := gobDecoder.Decode(&msg)
			checkError(err)

			if msg.MsgType == GETFILE {
				charXML := getCharacterXMLFromFile(msg.getMessage())
				err = gobEncoder.Encode(*charXML)
				checkError(err)
			} else {
				var char CharacterXML
				err := gobDecoder.Decode(&char)
				checkError(err)
				saveCharacterFile(&char)
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

func getCharacterXMLFromFile(charName string) *CharacterXML {

	fmt.Println("looking for : " + charName)
	xmlFile, err := os.Open("Characters/" + charName + ".xml")
	checkError(err)

	XMLdata, _ := ioutil.ReadAll(xmlFile)

	var charData CharacterXML
	xml.Unmarshal(XMLdata, &charData)
	xmlFile.Close()
	return &charData
}

func saveCharacterFile(char *CharacterXML) {
	fmt.Println("Saving char: ", char)
	file, err := os.Create("Characters/" + char.Name + ".xml")
	checkError(err)
	defer file.Close()

	enc := xml.NewEncoder(file)
	enc.Indent(" ", " ")

	err = enc.Encode(char)
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
			gob.NewEncoder(myConn).Encode(newServerMessageTypeS(REDIRECT, newAddress))
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
