package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

func main() {
	config() //读取执行参数
	c := New(ClipBoardOption{
		Mode: ClipBoardEnum(mode),
	})
	go c.Init().Run()
	fmt.Println("开始监听剪贴板")
	select {}
}

var (
	ip   string
	port int16
	mode string
)

func config() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "剪贴板配置参数",
		Long:  "剪贴板配置参数",
		Args:  cobra.MinimumNArgs(1),
	}
	configCmd.Flags().StringVarP(&ip, "ip", "i", "none", "服务器IP")
	configCmd.Flags().Int16VarP(&port, "port", "p", 0, "服务器端口")
	configCmd.Flags().StringVarP(&mode, "mode", "m", "server", "客户端类型: client or server")
	err := configCmd.Execute()
	if err != nil {
		panic(err)
	}
}

type ClipBoardEnum string

var (
	Client ClipBoardEnum = "client"
	Server ClipBoardEnum = "server"
)

type ClipBoardOption struct {
	Mode ClipBoardEnum
}
type ClipBoard struct {
	opt ClipBoardOption
	pub func([]byte)
}

func New(opt ClipBoardOption) *ClipBoard {
	c := &ClipBoard{
		opt: opt,
	}

	return c
}

func (c *ClipBoard) Init() *ClipBoard {
	switch c.opt.Mode {
	case Client:
		go c.client().run(c)
	case Server:
		go c.server().run(c)
	default:
		panic("不支持的模式")
	}
	return c
}

func (c *ClipBoard) Run() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	go func() {
		ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
		for data := range ch {
			c.pub(data)
		}
	}()
}

func (c *ClipBoard) server() *ClipBoardServer {
	return &ClipBoardServer{
		Port: port,
	}
}
func (c *ClipBoard) client() *ClipBoardClient {
	return &ClipBoardClient{
		ServerIP: ip,
		Port:     port,
	}
}

type protoc struct {
	Prefix string
}

// @jmrxlen(data)_xxxxxx@jmrxlen(data)_xxxxxx
func (p *protoc) read(buffer io.Reader) ([]byte, error) {
	buf := bufio.NewReader(buffer)
	for {
		b0, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if b0 != byte('@') {
			continue
		}
		b1 := make([]byte, len([]byte(p.Prefix)))
		n, err := buf.Read(b1)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal([]byte(p.Prefix), b1[:n]) {
			continue
		}
		b2, err := buf.ReadBytes(byte('_'))
		if err != nil {
			return nil, err
		}
		lenData := bytes.TrimSuffix(b2, []byte{byte('_')})
		dataLen, err := strconv.Atoi(string(lenData))
		if err != nil {
			return nil, err
		}
		body := make([]byte, dataLen)
		n, err = buf.Read(body)
		if err != nil {
			return nil, err
		}
		if n != dataLen {
			return nil, errors.New("协议错误")
		}
		return body, nil
	}
}

func (p *protoc) pkg(data []byte) []byte {
	buf := append([]byte{}, []byte(p.Prefix)...)
	buf = append(buf, []byte(strconv.Itoa(len(data)))...)
	buf = append(buf, byte('_'))
	buf = append(buf, data...)
	return buf
}

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
	addr := fmt.Sprintf(`127.0.0.1:%d`, c.Port)
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
		fmt.Println("客户端:", conn.RemoteAddr().String(), "连接成功")
		go c.connHandler(conn)
	}
}

func (c *ClipBoardServer) connHandler(conn net.Conn) {
	defer func() {
		conn.Close()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	key := conn.RemoteAddr().String()
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
		<-clipboard.Write(clipboard.FmtText, body)
	}
}

func (c *ClipBoardServer) publish(data []byte) {
	c.connMap.Range(func(key, value any) bool {
		protoc := protoc{
			Prefix: "@jmrx#@!%",
		}
		_, err := value.(net.Conn).Write(protoc.pkg(data))
		if err != nil {
			fmt.Println(err)
		}
		return true
	})
}

type ClipBoardClient struct {
	cb       *ClipBoard
	ServerIP string
	Port     int16
	conn     net.Conn
}

func (c *ClipBoardClient) run(cb *ClipBoard) {
	cb.pub = c.publish
	for {
		err := c.connect()
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("连接成功")
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
			fmt.Println(err)
		}
	}()
	for {
		pro := protoc{
			Prefix: "@jmrx#@!%",
		}
		body, err := pro.read(c.conn)
		if err != nil {
			panic(err)
		}
		<-clipboard.Write(clipboard.FmtText, body)
	}
}

func (c *ClipBoardClient) publish(data []byte) {
	protoc := protoc{
		Prefix: "@jmrx#@!%",
	}
	_, err := c.conn.Write(protoc.pkg(data))
	if err != nil {
		fmt.Println(err)
	}
}
