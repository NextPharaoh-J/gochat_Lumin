package connect

import (
	"github.com/gorilla/websocket"
	"gochat_my/proto"
	"net"
)

// user Connect session
type Channel struct {
	Room      *Room
	Next      *Channel
	Prev      *Channel
	broadcast chan *proto.Msg
	userId    int
	conn      *websocket.Conn
	connTCP   *net.TCPConn
}

func NewChannel(size int) (c *Channel) {
	return &Channel{
		broadcast: make(chan *proto.Msg, size),
	}
}

func (c *Channel) Push(msg *proto.Msg) (err error) {
	select {
	case c.broadcast <- msg:
	default:
	}
	return
}
