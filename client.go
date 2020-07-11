package main

import (
	"bytes"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	// 信息寫到Client最大等候時間
	writeWait = 10 * time.Second

	// Client傳信息過來最大等候時間
	pongWait = 60 * time.Second

	// 每pingPeriod去ping一次client，如果沒反應就斷掉連線
	pingPeriod = (pongWait * 9) / 10
)

// Client client的連線
type Client struct {
	UUID string
	conn *websocket.Conn
	hub  *Hub
	send chan []byte
}

// ReceiveClientMsgProcess 接收Client的信息
func (c *Client) ReceiveClientMsgProcess() {
	defer c.closeConn()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, msg, err := c.conn.ReadMessage()

		// 有錯就中斷連線
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Infof("%v", err)
			}
			break
		}

		// 格式處理
		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		// 同步把信息都推同一個hub上面的client
		c.hub.broadcast <- msg
	}
}

// SendClientMsgProcess 傳信息到Client
func (c *Client) SendClientMsgProcess() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.closeConn()
		ticker.Stop()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// channel被關掉的情況
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 傳Message
			err := c.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// ping有錯就斷線
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 連線中止
func (c *Client) closeConn() {
	c.hub.unregister <- c
	c.conn.Close()
}
