package clipboardshare

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type ClipBoardServer struct {
	Port    int16
	connMap sync.Map
	pro     protoc
	cb      *ClipBoard
	source  string
}

func (c *ClipBoardServer) showLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			log.Println("局域网ip: ", ip.String(), "mac: ", iface.HardwareAddr.String())
		}
	}
	return "127.0.0.1"
}

func (c *ClipBoardServer) listen() {
	addr := fmt.Sprintf(`0.0.0.0:%d`, c.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Println("开始监听,地址:", addr)
	c.showLocalIP()
	// go c.heart(time.Second)
	for {
		co, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go c.connHandler(co)
	}
}

func (c *ClipBoardServer) connHandler(co net.Conn) {
	log.Println("客户端:", co.RemoteAddr().String(), "连接成功")
	key := co.RemoteAddr().String()
	defer func() {
		co.Close()
		c.connMap.Delete(key)
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	client, ok := c.connMap.Load(key)
	if ok {
		client.(net.Conn).Close()
	}
	c.connMap.Store(key, co)

	for {
		w := bytes.NewBuffer(nil)
		err := c.pro.read(co, w)
		if err != nil {
			panic(err)
		}
		c.source = co.RemoteAddr().String()
		c.checkData(w.Bytes())
	}
}

func (c *ClipBoardServer) checkData(data []byte) {
	fmt.Println("[server]接受数据:", string(data))
	c.publish(data)
}

func (c *ClipBoardServer) heart(d time.Duration) {
	t := time.NewTicker(d)
	for _ = range t.C {
		c.connMap.Range(func(key, value any) bool {
			err := c.pro.w([]byte("heart"), value.(net.Conn))
			if err != nil {
				log.Println(err)
				c.connMap.Delete(key)
				return true
			}
			return true
		})
	}
}

func (c *ClipBoardServer) publish(data []byte) {
	c.connMap.Range(func(key, value any) bool {
		if key == c.source {
			return true
		}
		log.Println("[server]发送数据->", key, ":", string(data))
		err := c.pro.write(protocFrame{
			Type:     dataRaw,
			DataType: txt,
			Data:     data,
		}, value.(net.Conn))
		if err != nil {
			log.Println(err)
			c.connMap.Delete(key)
			return true
		}
		return true
	})
}

func (c *ClipBoardServer) connectSelf() {
	c.cb.client().register(c.cb).run()
}
