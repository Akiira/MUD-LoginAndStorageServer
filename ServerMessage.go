// ServerMessage
package main

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
	"strings"
)

const (
	ERROR    = 0
	REDIRECT = 1
	GETFILE  = 2
	SAVEFILE = 3
	GAMEPLAY = 4
	PING     = 5
)

type ServerMessage struct {
	Value   []FormattedString
	MsgType int
}

func newServerMessageFS(msgs []FormattedString) ServerMessage {
	return ServerMessage{MsgType: GAMEPLAY, Value: msgs}
}

func newServerMessageS(msg string) ServerMessage {
	return ServerMessage{MsgType: GAMEPLAY, Value: newFormattedStringSplice(msg)}
}

func newServerMessageTypeFS(typeOfMsg int, msgs []FormattedString) ServerMessage {
	return ServerMessage{MsgType: typeOfMsg, Value: msgs}
}

func newServerMessageTypeS(typeOfMsg int, msg string) ServerMessage {
	return ServerMessage{MsgType: typeOfMsg, Value: newFormattedStringSplice(msg)}
}

func newMessageWithRaces() ServerMessage {
	return newServerMessageS(fmt.Sprintf("\t%10s\t%10s\t%10s", "Elf", "Human", "Dwarf"))
}

func newMessageWithClasses() ServerMessage {
	return newServerMessageS(fmt.Sprintf("\t%10s\t%s", "Fighter", "Wizard"))
}

func NewMessageWithStats(stats []int) ServerMessage {
	fsc := newFormattedStringCollection()

	fsc.addMessage(ct.Green, "Stats\n-----------------------------------------\n")
	fsc.addMessage2(fmt.Sprintf("\tStrength: %2d", stats[0]))
	fsc.addMessage2(fmt.Sprintf("\tConstitution: %2d", stats[1]))
	fsc.addMessage2(fmt.Sprintf("\tDexterity: %2d", stats[2]))
	fsc.addMessage2(fmt.Sprintf("\tWisdom: %2d", stats[3]))
	fsc.addMessage2(fmt.Sprintf("\tCharisma: %2d", stats[4]))
	fsc.addMessage2(fmt.Sprintf("\tInteligence: %d2\n", stats[5]))

	return newServerMessageFS(fsc.fmtedStrings)
}

func (msg *ServerMessage) getMessage() string {
	if len(msg.Value) <= 0 {
		return ""
	}
	return msg.Value[0].Value
}

func (msg *ServerMessage) isError() bool {
	if len(msg.Value) == 0 {
		return false
	}

	return (strings.Split(msg.Value[0].Value, " ")[0] == "error")
}
