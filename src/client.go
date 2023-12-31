package clipboardshare

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

type ClipBoardClient struct {
	ServerIP      string
	Port          int16
	conn          net.Conn
	pro           protoc
	isReceiveData bool
}

func (c *ClipBoardClient) register(cb *ClipBoard) *ClipBoardClient {
	cb.pub = c.publish
	return c
}

func (c *ClipBoardClient) run() {
	for {
		err := c.connect()
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}
		log.Println("连接成功")
		c.connHandler()
	}
}

func (c *ClipBoardClient) connect() error {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.ServerIP == "" {
		c.ServerIP = "0.0.0.0"
	}
	addr := fmt.Sprintf(`%s:%d`, c.ServerIP, c.Port)
	co, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.conn = co
	return nil
}

func (c *ClipBoardClient) connHandler() {
	defer func() {
		if c.conn != nil {
			c.conn.Close()
		}
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	for {
		w := bytes.NewBuffer(nil)
		err := c.pro.read(c.conn, w)
		if err != nil {
			panic(err)
		}
		c.checkData(w.Bytes())
	}
}

func (c *ClipBoardClient) checkData(data []byte) {
	fmt.Println("[client]接收数据:", string(data))
	c.isReceiveData = true
	clipboardWrite(data)
}

func (c *ClipBoardClient) publish(data []byte) {
	if c.isReceiveData {
		c.isReceiveData = false
		return
	}
	if c.conn != nil {
		log.Println("[client]发送数据->", c.conn.RemoteAddr().String(), ":", string(data))
		err := c.pro.write(protocFrame{
			Type:     dataRaw,
			DataType: txt,
			Data:     data,
		}, c.conn)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
