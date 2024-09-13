package connect

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gochat_my/proto"
	"gochat_my/tools"
	"time"
)

type Server struct {
	Buckets   []*Bucket
	Options   ServerOptions
	bucketIdx uint32
	operator  Operator
}

type ServerOptions struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int
}

func NewServer(b []*Bucket, o Operator, opts ServerOptions) *Server {
	return &Server{
		Buckets:   b,
		Options:   opts,
		bucketIdx: uint32(len(b)),
		operator:  o,
	}
}

func (s *Server) Bucket(userId int) *Bucket {
	userIdStr := fmt.Sprintf("%d", userId)
	idx := tools.CityHash32([]byte(userIdStr), uint32(len(userIdStr))) % s.bucketIdx
	return s.Buckets[idx]
}

func (s *Server) writePump(ch *Channel, c *Connect) {
	// pingPeriod default eq 54s
	ticker := time.NewTicker(s.Options.PingPeriod)
	defer func() {
		ticker.Stop()
		ch.conn.Close()
	}()
	for {
		select {
		case message, ok := <-ch.broadcast:
			ch.conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			if !ok {
				logrus.Warn("SetWriteDeadline Failed")
				ch.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := ch.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logrus.Warn("ch.conn.NextWriter Failed : %s", err.Error())
				return
			}
			logrus.Infof("message write body : %s ", message.Body)
			w.Write(message.Body)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// heartbeat ,if ping error will exit and close current websocket conn
			ch.conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			logrus.Infof("websocket.PingMessage : %v", websocket.PingMessage)
			if err := ch.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Server) readPump(ch *Channel, c *Connect) {
	defer func() {
		logrus.Infof("start exec disConnect")
		if ch.Room == nil || ch.userId == 0 {
			logrus.Infof("roomId or userId equal to 0")
			ch.conn.Close()
			return
		}
		logrus.Infof("exec disConnect ...")
		disConnectRequest := &proto.DisConnectRequest{
			RoomId: ch.Room.Id,
			UserId: ch.userId,
		}
		s.Bucket(ch.userId).DeleteChannel(ch)
		if err := s.operator.DisConnect(disConnectRequest); err != nil {
			logrus.Warnf("DisConnect err : %s ", err.Error())
		}
		ch.conn.Close()
	}()
	ch.conn.SetReadLimit(s.Options.MaxMessageSize)
	ch.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
	ch.conn.SetPongHandler(func(string) error {
		ch.conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
		return nil
	})
	for {
		_, message, err := ch.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("readPump ReadMessage err : %s ", err.Error())
				return
			}
		}
		if message == nil {
			return
		}
		var connReq *proto.ConnectRequest
		logrus.Infof("get a message : %s ", message)
		if err = json.Unmarshal(message, &connReq); err != nil {
			logrus.Errorf("message struct %+v ", connReq) // 日志结构输出键值对
		}
		if connReq == nil || connReq.AuthToken == "" {
			logrus.Errorf("Connect no authToken")
			return
		}
		connReq.ServerId = c.ServerId
		userId, err := s.operator.Connect(connReq)
		if err != nil {
			logrus.Errorf("Connect err : %s ", err.Error())
			return
		}
		if userId == 0 {
			logrus.Errorf("Invalid AuthToken ,userId empty")
			return
		}
		logrus.Infof("websocket rpc call return userId:%d , RoomId:%d", userId, connReq.RoomId)
		b := s.Bucket(userId)
		// insert into Bucket
		err = b.Put(userId, connReq.RoomId, ch)
		if err != nil {
			logrus.Errorf("channel close err : %s ", err.Error())
			ch.conn.Close()
		}
	}

}
