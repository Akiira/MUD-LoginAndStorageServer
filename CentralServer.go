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
	"sync"
)

const (
	PASSWORD = 0
	ADDRESS  = 1
)

var servers map[string]string
var fileSystemMutex sync.Mutex
var passwordPath string = "Characters/Passwords/"

func main() {

	gob.Register(WeaponXML{})
	gob.Register(ArmourXML{})
	gob.Register(ArmourSetXML{})
	gob.Register(ItemXML{})

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
			gobDecoder := gob.NewDecoder(conn)
			gobEncoder := gob.NewEncoder(conn)
			err := gobDecoder.Decode(&msg)
			checkError(err)

			if msg.MsgType == GETFILE {
				charXML := getCharacterXMLFromFile(msg.getMessage())
				fmt.Println("Sending char: ", *charXML)
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

func HandleLoginClient(myConn net.Conn) {
	var clientsMsg ClientMessage
	var servMsg ServerMessage
	defer myConn.Close()

	err := gob.NewDecoder(myConn).Decode(&clientsMsg)
	checkError(err)

	if DoesCharacterExist(clientsMsg.getUsername()) {

		if GetCharactersPassword(clientsMsg.getPassword()) == clientsMsg.getPassword() {
			servMsg = newServerMessageTypeS(REDIRECT, GetCharactersWorld(clientsMsg.getUsername()))
		} else {
			servMsg = newServerMessageTypeS(ERROR, "Incorrect password, closing connection.")
		}
	} else {
		servMsg = newServerMessageTypeS(ERROR, "Character does not exist, closing connection.")
	}

	err = gob.NewEncoder(myConn).Encode(servMsg)
	checkError(err)
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

func getCharacterXMLFromFile(charName string) *CharacterXML {
	fmt.Println("looking for : " + charName)
	xmlFile, err := os.Open("Characters/" + charName + ".xml")
	checkError(err)
	defer xmlFile.Close()

	XMLdata, err := ioutil.ReadAll(xmlFile)
	checkError(err)

	var charData CharacterXML
	err = xml.Unmarshal(XMLdata, &charData)
	checkError(err)

	return &charData
}

func saveCharacterFile(char *CharacterXML) {
	fmt.Println("Saving char: ", char)
	file, err := os.Create("Characters/" + char.Name + ".xml")
	checkError(err)
	defer file.Close()

	enc := xml.NewEncoder(file)
	enc.Indent(" ", "\t")

	err = enc.Encode(char)
	checkError(err)

	UpdatePasswordFile(char.Name, GetCharactersPassword(char.Name), char.CurrentWorld)
}

func GetCharactersPassword(name string) (password string) {
	pwFile, err := os.Open(passwordPath + name + ".txt")
	checkError(err)
	defer pwFile.Close()

	reader := bufio.NewReader(pwFile)
	line, _, err := reader.ReadLine()
	checkError(err)
	pwAndWorld := strings.Split(string(line), " ")

	return pwAndWorld[PASSWORD]
}

func GetCharactersWorld(name string) string {
	pwFile, err := os.Open(passwordPath + name + ".txt")
	checkError(err)
	defer pwFile.Close()

	reader := bufio.NewReader(pwFile)
	line, _, err := reader.ReadLine()
	checkError(err)
	pwAndWorld := strings.Split(string(line), " ")

	return pwAndWorld[ADDRESS]
}

func DoesCharacterExist(name string) (found bool) {
	if _, err := os.Stat("Characters/" + name + ".xml"); err != nil {
		found = false
	} else {
		found = true
	}

	return found
}

func UpdatePasswordFile(name, password, world string) {
	pwFile, err := os.Create(passwordPath + name + ".txt")
	checkError(err)
	defer pwFile.Close()

	_, err = pwFile.Write([]byte(password + " " + world))
	checkError(err)
}

func CreateNewCharacter(encder *gob.Encoder, decder *gob.Decoder) {
	fileSystemMutex.Lock()
	defer fileSystemMutex.Unlock()

	var msg ClientMessage
	var charData CharacterXML

	//ask for name
	encder.Encode(newServerMessageS("Enter a name for your adventurer.\n"))

	//get name
	decder.Decode(&msg)

	//check name is not taken
	if DoesCharacterExist(msg.getUsername()) == false {
		//break
	}

	os.Create("Characters/Passwords/" + msg.getUsername() + ".txt")

	//ask for password
	encder.Encode(newServerMessageS("Enter a name for your adventurer.\n"))

	//get password
	decder.Decode(&msg)

	//display races
	encder.Encode(newMessageWithRaces()) //TODO

	//get selection
	decder.Decode(&msg)

	//display classes
	encder.Encode(newMessageWithClasses()) //TODO

	//get selection
	decder.Decode(&msg)

	//roll stats
	//display stats
	//reroll if desired

	//when accepted save to xml file.
	saveCharacterFile(&charData)

	//save the password file
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
