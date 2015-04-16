// ServerMessage
package main

import (
	"fmt"
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
