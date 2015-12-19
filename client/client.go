package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//拼接连接信息
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	//连接服务器
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			//堵塞监听信息
			_, message, err := c.ReadMessage()
			if err != nil {
				//输出错误信息
				log.Println("read:", err)
				return
			}
			//输出正常信息
			log.Printf("recv: %s", message)
		}
	}()

	//用户认证
	auth_str := "username=client&token=123456"
	auth_err := c.WriteMessage(websocket.TextMessage, []byte(auth_str))
	if auth_err != nil {
		log.Println("write:", auth_err)
		return
	}

	//加入定时器，定时发送消息

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		//向服务器发送消息
		case <-ticker.C:
		//断开连接
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}
