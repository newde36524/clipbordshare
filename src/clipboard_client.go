package clipboardshare

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

type ClipBoardClient struct {
	ServerIP string
	Port     int16
	conn     net.Conn
	pro      protoc
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
	clipboardWrite(data)
}

func (c *ClipBoardClient) publish(data []byte) {
	if c.conn != nil {
		err := c.pro.write(protocFrame{
			Type:     dataRaw,
			DataType: txt,
			Data:     data,
		}, c.conn)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("发送数据->", c.conn.RemoteAddr().String(), ":", string(data))
	}
}
