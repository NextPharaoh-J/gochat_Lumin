package connect

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gochat_my/config"
	"net/http"
)

func (c *Connect) InitWebSocket() (err error) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c.serveWS(DefaultServer, w, r)
	})
	err = http.ListenAndServe(config.Conf.Connect.ConnectWebsocket.Bind, nil)
	return
}

func (c *Connect) serveWS(server *Server, w http.ResponseWriter, r *http.Request) {
	var upGrader = websocket.Upgrader{
		ReadBufferSize:  server.Options.ReadBufferSize,
		WriteBufferSize: server.Options.WriteBufferSize,
	}
	// cross origin domain support
	upGrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upGrader.Upgrade(w, r, nil)

	if err != nil {
		logrus.Errorf("serverWS err : %s", err.Error())
		return
	}
	var ch *Channel
	ch = NewChannel(server.Options.BroadcastSize) //512
	ch.conn = conn
	// send data to websocket conn
	go server.writePump(ch, c)
	// get data from websocket conn
	go server.readPump(ch, c)

}
