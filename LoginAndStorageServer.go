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
	"time"
)

const (
	PASSWORD   = 0
	SRVER_NAME = 1
)

var (
	servers         map[string]string = make(map[string]string)
	worldServers    map[string]string = make(map[string]string)
	passwordPath    string            = "Characters/Passwords/"
	fileSystemMutex sync.Mutex
)

func main() {

	gob.Register(WeaponXML{})
	gob.Register(ArmourXML{})
	gob.Register(ArmourSetXML{})
	gob.Register(ItemXML{})

	readServerList()
	go RunStorageServer()
	go RunNewCharacterServer()
	go RunLoginServer()
	go PeriodicHeartbeat()
	getInputFromUser()
}

// =======  HEART BEAT FUNCTIONS

func PeriodicHeartbeat() {
	for {
		RunHeartbeat()
		time.Sleep(10 * time.Second)
	}
}

func RunHeartbeat() {

	for name, address := range worldServers {
		err := GetHeartbeat(address)

		if err != nil {
			fmt.Println("error: cannot connect " + name + "\n")
			//TODO here we could add code to try and restart that world server or somthing.
			continue
		} else {
			fmt.Println("Heartbeat received from: ", name)
		}
	}
}

func GetHeartbeat(serverAddress string) error {
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return err
	}

	err = gob.NewEncoder(conn).Encode(ClientMessage{Command: "heartbeat"})
	if err != nil {
		return err
	}

	return gob.NewDecoder(conn).Decode(&ServerMessage{})
}

func getInputFromUser() {

	var input string
	for {

		_, err := fmt.Scan(&input)
		checkError(err, false)
		input = strings.TrimSpace(input)

		if input == "exit" {
			os.Exit(1)
		} else if input == "refreshserver" {
			readServerList()
			updateServerListToServers()
		}
	}
}

// =======  SERVER FUNCTIONS

func readServerList() {
	servers = make(map[string]string)
	file, err := os.Open("serverConfig/serverList.txt")
	if err != nil {
		log.Fatal(err, false)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		readData := strings.Fields(scanner.Text())
		fmt.Println(readData[0], " ", readData[1])
		servers[readData[0]] = readData[1]

		if strings.HasPrefix(readData[0], "world") {
			worldServers[readData[0]] = readData[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err, false)
	}
}

func setUpServerWithAddress(addr string) *net.TCPListener {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	checkError(err, false)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err, false)
	return listener
}

func updateServerListToServers() {

	var updatedAddresses string
	for name, address := range servers {
		updatedAddresses += name + " " + address + "\n"
	}

	for name, address := range servers {

		if name == "central" || name == "characterStorage" || name == "newChar" {
			continue
		} else {

			conn, err := net.Dial("tcp", address)
			checkError(err, false)
			encoder := gob.NewEncoder(conn)

			if err != nil {
				fmt.Println("error: cannot connect " + name + "\n")
			} else {
				encoder.Encode(newClientMessage("refreshserver", updatedAddresses))
			}

			err = conn.Close()
			checkError(err, false)
		}
	}

}

func RunStorageServer() {
	listener := setUpServerWithAddress(servers["characterStorage"])
	fmt.Println("Storage Server: i'm waiting")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Storage Server:Connection established")
		HandleStorageConnection(conn)
	}
	fmt.Println("Storage Server: i'm shutting down")
}

func HandleStorageConnection(conn net.Conn) {
	fileSystemMutex.Lock()
	defer fileSystemMutex.Unlock()
	defer conn.Close()

	gobDecoder := gob.NewDecoder(conn)

	var msg ServerMessage
	err := gobDecoder.Decode(&msg)
	checkError(err, false)

	if msg.MsgType == GETFILE {
		charXML := GetCharacterXML(msg.getMessage())
		err = gob.NewEncoder(conn).Encode(*charXML)
		checkError(err, false)
	} else {
		var char CharacterXML
		err := gobDecoder.Decode(&char)
		checkError(err, false)
		SaveCharacterData(&char)
	}
}

func RunLoginServer() {
	listener := setUpServerWithAddress(servers["central"])
	fmt.Println("Login Server: i'm waiting")

	for {
		conn, err := listener.Accept()
		checkError(err, false)
		if err == nil {
			fmt.Println("Login Server:Connection established")
			go HandleLogin(conn)
		}
	}
	fmt.Println("Login Server: i'm shutting down")
}

func HandleLogin(myConn net.Conn) {
	fileSystemMutex.Lock()
	defer fileSystemMutex.Unlock()
	defer myConn.Close()

	var clientsMsg ClientMessage
	var servMsg ServerMessage

	err := gob.NewDecoder(myConn).Decode(&clientsMsg)
	checkError(err, false)

	if CharacterExists(clientsMsg.getUsername()) {
		if GetCharactersPassword(clientsMsg.getUsername()) == clientsMsg.getPassword() {
			servMsg = newServerMessageTypeS(REDIRECT, GetCharactersWorld(clientsMsg.getUsername()))
		} else {
			servMsg = newServerMessageTypeS(ERROR, "Incorrect password, closing connection.")
		}
	} else {
		servMsg = newServerMessageTypeS(ERROR, "Character does not exist, closing connection.")
	}

	err = gob.NewEncoder(myConn).Encode(servMsg)
	checkError(err, false)
}

func RunNewCharacterServer() {
	listener := setUpServerWithAddress(servers["newChar"])
	fmt.Println("New Character Server up.")

	for {
		conn, err := listener.Accept()
		checkError(err, false)
		fmt.Println("NewConnection in NewCharServer")
		CreateNewCharacter(gob.NewEncoder(conn), gob.NewDecoder(conn))
		conn.Close()
	}
}

// =======  CHARACTER STORAGE AND VALIDATION FUNCTIONS

func GetCharacterXML(charName string) *CharacterXML {
	fmt.Println("looking for : " + charName)
	xmlFile, err := os.Open("Characters/" + charName + ".xml")
	checkError(err, false)
	defer xmlFile.Close()

	XMLdata, err := ioutil.ReadAll(xmlFile)
	checkError(err, false)

	var charData CharacterXML
	err = xml.Unmarshal(XMLdata, &charData)
	checkError(err, false)

	return &charData
}

func SaveCharacterData(char *CharacterXML) {
	fmt.Println("Saving char: ", char)
	file, err := os.Create("Characters/" + char.Name + ".xml")
	checkError(err, false)
	defer file.Close()

	enc := xml.NewEncoder(file)
	enc.Indent(" ", "\t")

	err = enc.Encode(char)
	checkError(err, false)

	UpdatePasswordFile(char.Name, GetCharactersPassword(char.Name), char.CurrentWorld)
}

func GetCharactersPassword(name string) (password string) {
	pwFile, err := os.Open(passwordPath + name + ".txt")
	checkErrorWithMessage(err, "Failed to open password file for: "+name)
	defer pwFile.Close()

	reader := bufio.NewReader(pwFile)
	line, _, err := reader.ReadLine()
	checkError(err, false)
	pwAndWorld := strings.Split(string(line), " ")

	return pwAndWorld[PASSWORD]
}

func GetCharactersWorld(name string) string {
	pwFile, err := os.Open(passwordPath + name + ".txt")
	checkErrorWithMessage(err, "Failed to open password file to get world id for character: "+name)
	defer pwFile.Close()

	reader := bufio.NewReader(pwFile)
	line, _, err := reader.ReadLine()
	checkError(err, false)
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
	checkError(err, false)

	fmt.Println("Password file for ", name, " updated.")
}

// =======  CHARACTER CREATION FUNCTIONS

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
	checkError(err, false)
	err = decder.Decode(&msg)
	checkError(err, false)
	password := msg.Value

	//display races
	err = encder.Encode(newMessageWithRaces())
	checkError(err, false)
	err = decder.Decode(&msg)
	checkError(err, false)
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
	SaveCharacterData(&charData)

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

// =======  ERROR CHECKING FUNCTIONS

func checkErrorWithMessage(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		fmt.Println(msg)
		os.Exit(1)
	}
}

func checkError(err error, exit bool) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())

		if exit {
			os.Exit(1)
		}
	}
}
