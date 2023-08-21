package clipboardshare

import (
	"bytes"
	"context"
	"log"

	"golang.design/x/clipboard"
)

type ClipBoardMode string

const (
	Client ClipBoardMode = "client"
	Server ClipBoardMode = "server"
)

type ClipBoardOption struct {
	Mode ClipBoardMode
	IP   string
	Port int16
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
	if err := clipboard.Init(); err != nil {
		panic(err)
	}
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	lastData := make([]byte, 0)
	for data := range ch {
		if !bytes.Equal(lastData, data) {
			log.Println("剪贴板数据:", string(data))
			lastData = data
			c.pub(data)
		}
	}
}

func (c *ClipBoard) server() *ClipBoardServer {
	return &ClipBoardServer{
		Port: c.opt.Port,
		pro: protoc{
			Prefix: "@jmrx#@!%",
		},
	}
}

func (c *ClipBoard) client() *ClipBoardClient {
	return &ClipBoardClient{
		ServerIP: c.opt.IP,
		Port:     c.opt.Port,
		pro: protoc{
			Prefix: "@jmrx#@!%",
		},
	}
}

func clipboardWrite(body []byte) {
	<-clipboard.Write(clipboard.FmtText, body)
	<-clipboard.Write(clipboard.FmtText, body)
}
