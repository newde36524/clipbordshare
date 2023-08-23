package clipboardshare

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
)

type ClipBoardServer struct {
	Port    int16
	connMap sync.Map
	pro     protoc
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

func (c *ClipBoardServer) register(cb *ClipBoard) *ClipBoardServer {
	cb.pub = c.publish
	return c
}

func (c *ClipBoardServer) listen() {
	addr := fmt.Sprintf(`0.0.0.0:%d`, c.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Println("开始监听,地址:", addr)
	c.showLocalIP()
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
		c.checkData(w.Bytes())
	}
}

func (c *ClipBoardServer) checkData(data []byte) {
	fmt.Println("接受数据:", string(data))
	clipboardWrite(data)
}

func (c *ClipBoardServer) publish(data []byte) {
	c.connMap.Range(func(key, value any) bool {
		err := c.pro.write(protocFrame{
			Type:     dataRaw,
			DataType: txt,
			Data:     data,
		}, value.(net.Conn))
		if err != nil {
			log.Println(err)
			return true
		}
		log.Println("发送数据->", key, ":", string(data))
		return true
	})
}
