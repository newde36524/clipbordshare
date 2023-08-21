package clipboardshare

import (
	"fmt"
	"net"
	"sync"
)

type ClipBoardServer struct {
	cb      *ClipBoard
	Port    int16
	connMap sync.Map
}

func (c *ClipBoardServer) showLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
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
			fmt.Println("局域网ip: ", ip.String(), "mac: ", iface.HardwareAddr.String())
		}
	}
	return "127.0.0.1"
}

func (c *ClipBoardServer) run(cb *ClipBoard) {
	cb.pub = c.publish
	addr := fmt.Sprintf(`0.0.0.0:%d`, c.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	fmt.Println("开始监听,地址:", addr)
	c.showLocalIP()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go c.connHandler(conn)
	}
}

func (c *ClipBoardServer) connHandler(conn net.Conn) {
	fmt.Println("客户端:", conn.RemoteAddr().String(), "连接成功")
	key := conn.RemoteAddr().String()
	defer func() {
		conn.Close()
		c.connMap.Delete(key)
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	client, ok := c.connMap.Load(key)
	if ok {
		client.(net.Conn).Close()
	}
	c.connMap.Store(key, conn)
	for {
		pro := protoc{
			Prefix: "@jmrx#@!%",
		}
		body, err := pro.read(conn)
		if err != nil {
			panic(err)
		}
		clipboardWrite(body)
		c.publish(body)
	}
}

func (c *ClipBoardServer) publish(data []byte) {
	c.connMap.Range(func(key, value any) bool {
		protoc := protoc{
			Prefix: "@jmrx#@!%",
		}
		pkg := protoc.pkg(data)
		for {
			n, err := value.(net.Conn).Write(pkg)
			if err != nil {
				fmt.Println(err)
				break
			}
			if n == 0 {
				fmt.Println("发送失败")
				break
			}
			if len(pkg) == n {
				fmt.Println("发送成功")
				break
			} else {
				pkg = pkg[n:]
			}
		}
		fmt.Println("发送数据->", key, ":", string(data))
		return true
	})
}