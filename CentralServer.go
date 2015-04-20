package main

import (
	"bufio"
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	PASSWORD   = 0
	SRVER_NAME = 1
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
	go RunNewCharacterServer()
	runClientServer()
}

func RunNewCharacterServer() {
	listener := setUpServerWithAddress(servers["newChar"])
	fmt.Println("\tNew Character Server up.")
	for {
		conn, err := listener.Accept()
		checkError(err)
		fmt.Println("\tNewConnection in NewCharServer")
		CreateNewCharacter(gob.NewEncoder(conn), gob.NewDecoder(conn))
		conn.Close()
	}
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
				charXML := GetCharacterXML(msg.getMessage())
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

	if CharacterExists(clientsMsg.getUsername()) {
		if GetCharactersPassword(clientsMsg.getUsername()) == clientsMsg.getPassword() {
			servMsg = newServerMessageTypeS(REDIRECT, GetCharactersWorld(clientsMsg.getUsername()))
		} else {
			servMsg = newServerMessageTypeS(ERROR, "Incorrect password, closing connection.")
		}
	} else {
		servMsg = newServerMessageTypeS(ERROR, "Character does not exist, closing connection.")
	}

	fmt.Println("\tSending message: ", servMsg.Value, ", ", servMsg.MsgType)
	err = gob.NewEncoder(myConn).Encode(servMsg)
	checkError(err)
}

func GetCharacterXML(charName string) *CharacterXML {
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
	checkErrorWithMessage(err, "Failed to open password file for: "+name)
	defer pwFile.Close()

	reader := bufio.NewReader(pwFile)
	line, _, err := reader.ReadLine()
	checkError(err)
	pwAndWorld := strings.Split(string(line), " ")

	return pwAndWorld[PASSWORD]
}

func GetCharactersWorld(name string) string {
	pwFile, err := os.Open(passwordPath + name + ".txt")
	checkErrorWithMessage(err, "Failed to open password file to get world id for character: "+name)
	defer pwFile.Close()

	reader := bufio.NewReader(pwFile)
	line, _, err := reader.ReadLine()
	checkError(err)
	pwAndWorld := strings.Split(string(line), " ")

	return servers[pwAndWorld[SRVER_NAME]]
}

func CharacterExists(name string) (found bool) {
	if _, err := os.Stat("Characters/" + name + ".xml"); err != nil {
		found = false
	} else {
		found = true
	}

	return found
}

func UpdatePasswordFile(name, password, world string) {
	pwFile, err := os.Create(passwordPath + name + ".txt")
	checkErrorWithMessage(err, "Failed to create password file for: "+name)
	defer pwFile.Close()

	_, err = pwFile.Write([]byte(password + " " + world))
	checkError(err)

	fmt.Println("Password file for ", name, " updated.breakSignal")
}

func CreateNewCharacter(encder *gob.Encoder, decder *gob.Decoder) {
	fileSystemMutex.Lock()
	defer fileSystemMutex.Unlock()

	var msg ClientMessage
	var charData CharacterXML
	charData.SetToDefaultValues()

	for {
		//ask for name
		err := encder.Encode(newServerMessageS("Enter a name for your adventurer.\n"))
		checkErrorWithMessage(err, "Send msg for name for adventurer in CreateNewChar.")
		err = decder.Decode(&msg)
		checkErrorWithMessage(err, "Reading name for adventurer in CreateNewChar.")
		charData.Name = msg.Value

		//check name is not taken
		if !CharacterExists(msg.getUsername()) {
			break
		} else {
			err := encder.Encode(newServerMessageS("That name is already taken.\n"))
			checkErrorWithMessage(err, "Send msg for name already exists.")
		}
	}

	//ask for password
	err := encder.Encode(newServerMessageS("Enter a password.\n"))
	checkError(err)
	err = decder.Decode(&msg)
	checkError(err)
	password := msg.Value

	//display races
	err = encder.Encode(newMessageWithRaces())
	checkError(err)
	err = decder.Decode(&msg)
	checkError(err)
	charData.Race = msg.Value //TODO chekc valid choice

	//display classes
	encder.Encode(newMessageWithClasses())
	decder.Decode(&msg)
	charData.Class = msg.Value //TODO chekc valid choice

	for {
		//roll stats
		stats := RollStats()

		//display stats
		encder.Encode(NewMessageWithStats(stats))

		//reroll if desired
		decder.Decode(&msg)
		if msg.Value != "reroll" {
			charData.SetStats(stats)
			break
		}
	}

	//save the password and character file
	UpdatePasswordFile(charData.Name, password, "world1")
	saveCharacterFile(&charData)

	encder.Encode(newServerMessageTypeS(EXIT, "Your character was created succesfully."+
		"You will now be redidrected to the login server. Press 'done' to continue.\n"))
}

func RollStats() []int {
	stats := make([]int, 6)

	for index, _ := range stats {
		stats[index] = RollD6() + RollD6() + RollD6()
	}

	return stats
}

func RollD6() int {
	return rand.Intn(6) + 1
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

func checkErrorWithMessage(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		fmt.Println(msg)
		os.Exit(1)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
