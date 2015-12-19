// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	//对方写入会话等待时间
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	//对方读取下次消息等待时间
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	//对方ping周期
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	//对方最大写入字节数
	// Maximum message size allowed from peer.
	maxMessageSize = 512

	//验证字符串
	authToken = "123456"
)

//服务器配置信息
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection 是websocket的conntion和hub的中间人
// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	//websocket的连接
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	//出站消息缓存通道
	send chan []byte

	//验证状态
	auth bool

	//验证状态
	username []byte
}

//读取connection中的数据导入到hub中，实则发广播消息
//服务器读取的所有客户端的发来的消息
// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump() {
	defer func() {
		h.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		mtype := 2 //用户消息
		text := string(message)
		reg := regexp.MustCompile(`=[^&]+`)
		s := reg.FindAllString(text, -1)

		//默认all
		if len(s) == 2 {
			fromuser := strings.Replace(s[0], "=", "", 1)
			token := strings.Replace(s[1], "=", "", 1)
			if token == authToken {
				c.username = []byte(fromuser)
				c.auth = true
				message = []byte(fromuser + " join")
				mtype = 1 //系统消息
			}
		}

		touser := []byte("all")
		reg2 := regexp.MustCompile(`^@.*? `)
		s2 := reg2.FindAllString(text, -1)
		if len(s2) == 1 {
			s2[0] = strings.Replace(s2[0], "@", "", 1)
			s2[0] = strings.Replace(s2[0], " ", "", 1)
			touser = []byte(s2[0])
		}

		if c.auth == true {
			t := time.Now().Unix()
			h.broadcast <- &tmessage{content: message, fromuser: c.username, touser: touser, mtype: mtype, createtime: time.Unix(t, 0).String()}
		}
	}
}

//给消息，指定消息类型和荷载
// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

//从hub到connection写数据
//服务器端发送消息给客户端
// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	//定时执行
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

//处理客户端对websocket请求
// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	//设定环境变量
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//初始化connection
	c := &connection{send: make(chan []byte, 256), ws: ws, auth: false}
	//加入注册通道，意思是只要连接的人都加入register通道
	h.register <- c
	go c.writePump() //服务器端发送消息给客户端
	c.readPump()     //服务器读取的所有客户端的发来的消息
}
