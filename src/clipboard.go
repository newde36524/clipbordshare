package clipboardshare

import (
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
	Mode     ClipBoardMode
	IP       string
	Port     int16
	Prefix   string
	PageSize int
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
		go c.client().register(c).run()
	case Server:
		srv := c.server().register(c)
		go srv.listen()
		go srv.connectSelf()
	default:
		panic("不支持的模式:" + c.opt.Mode)
	}
	return c
}

func (c *ClipBoard) Run() {
	if err := clipboard.Init(); err != nil {
		panic(err)
	}
	log.Println("开始监听剪贴板")
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	for data := range ch {
		if len(data) != 0 {
			log.Println("更新剪贴板数据:", string(data))
			c.pub(data)
		}
	}
}

func (c *ClipBoard) server() *ClipBoardServer {
	return &ClipBoardServer{
		Port: c.opt.Port,
		pro: protoc{
			Prefix:   c.opt.Prefix,
			PageSize: c.opt.PageSize,
		},
	}
}

func (c *ClipBoard) client() *ClipBoardClient {
	return &ClipBoardClient{
		ServerIP: c.opt.IP,
		Port:     c.opt.Port,
		pro: protoc{
			Prefix:   c.opt.Prefix,
			PageSize: c.opt.PageSize,
		},
	}
}

func clipboardWrite(body []byte) {
	clipboard.Write(clipboard.FmtText, body)
}
