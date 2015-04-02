// ServerMessage
package main

const (
	REDIRECT = 1
)

type ServerMessage struct {
	MsgType int
	Value   []FormattedString
}

func newServerMessage(typeOfMsg int, msg string) ServerMessage {
	return ServerMessage{MsgType: typeOfMsg, Value: newFormattedString(msg)}
}
