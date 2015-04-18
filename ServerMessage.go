// ServerMessage
package main

import (
	"fmt"
	"github.com/daviddengcn/go-colortext"
)

const (
	ERROR    = 0
	REDIRECT = 1
	GETFILE  = 2
	SAVEFILE = 3
	GAMEPLAY = 4
	PING     = 5
	EXIT     = 6
)

type ServerMessage struct {
	Value    []FormattedString
	MsgType  int
	CharInfo CharacterInfo
}

type CharacterInfo struct {
	Name      string
	CurrentHP int
	MaxHP     int
	Exp       int
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

func NewRedirectMsgS(msg string) ServerMessage {
	return ServerMessage{MsgType: REDIRECT, Value: newFormattedStringSplice(msg)}
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
	fsc.addMessage2(fmt.Sprintf("\tStrength: %2d\n", stats[0]))
	fsc.addMessage2(fmt.Sprintf("\tConstitution: %2d\n", stats[1]))
	fsc.addMessage2(fmt.Sprintf("\tDexterity: %2d\n", stats[2]))
	fsc.addMessage2(fmt.Sprintf("\tWisdom: %2d\n", stats[3]))
	fsc.addMessage2(fmt.Sprintf("\tCharisma: %2d\n", stats[4]))
	fsc.addMessage2(fmt.Sprintf("\tInteligence: %2d\n", stats[5]))

	fsc.addMessage(ct.Green, "\nType 'reroll' to get new stats or 'keep' to continue.\n")

	return newServerMessageFS(fsc.fmtedStrings)
}

func (msg *ServerMessage) getFormattedCharInfo() []FormattedString {
	fs := newFormattedStringCollection()
	fs.addMessage(ct.Red, fmt.Sprintf("\n%d/%d", msg.getCurrentHP(), msg.getMaxHP()))
	fs.addMessage2("|")
	fs.addMessage(ct.Green, fmt.Sprintf("%d", msg.CharInfo.Exp))
	fs.addMessage2("|")
	fs.addMessage(ct.Blue, msg.CharInfo.Name)
	fs.addMessage2("> ")
	return fs.fmtedStrings
}

func (msg *ServerMessage) getCurrentHP() int {
	return msg.CharInfo.CurrentHP
}

func (msg *ServerMessage) getMaxHP() int {
	return msg.CharInfo.MaxHP
}

func (msg *ServerMessage) getMessage() string {
	if len(msg.Value) <= 0 {
		return ""
	}
	return msg.Value[0].Value
}
