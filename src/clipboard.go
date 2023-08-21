package clipboardshare

import (
	"context"

	"golang.design/x/clipboard"
)

type ClipBoardEnum string

var (
	Client ClipBoardEnum = "client"
	Server ClipBoardEnum = "server"
)

type ClipBoardOption struct {
	Mode ClipBoardEnum
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
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for data := range ch {
		c.pub(data)
	}
}

func (c *ClipBoard) server() *ClipBoardServer {
	return &ClipBoardServer{
		Port: c.opt.Port,
	}
}

func (c *ClipBoard) client() *ClipBoardClient {
	return &ClipBoardClient{
		ServerIP: c.opt.IP,
		Port:     c.opt.Port,
	}
}

func clipboardWrite(body []byte) {
	<-clipboard.Write(clipboard.FmtText, body)
}
