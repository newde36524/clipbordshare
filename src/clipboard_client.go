package clipboardshare

import (
	"fmt"
	"log"
	"net"
	"time"
)

type ClipBoardClient struct {
	cb       *ClipBoard
	ServerIP string
	Port     int16
	conn     net.Conn
	pro      protoc
}

func (c *ClipBoardClient) run(cb *ClipBoard) {
	cb.pub = c.publish
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
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.conn = conn
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
		body, err := c.pro.read(c.conn)
		if err != nil {
			panic(err)
		}
		clipboardWrite(body)
	}
}

func (c *ClipBoardClient) publish(data []byte) {
	if c.conn != nil {
		pkg := c.pro.pkg(data)
		for {
			n, err := c.conn.Write(pkg)
			if err != nil {
				log.Println(err)
				break
			}
			if n == 0 {
				log.Println("发送失败")
				break
			}
			if len(pkg) == n {
				log.Println("发送成功")
				break
			} else {
				pkg = pkg[n:]
			}
		}
		log.Println("发送数据->", c.conn.RemoteAddr().String(), ":", string(data))
	}
}
