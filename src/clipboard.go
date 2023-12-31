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
	pro protoc
}

func New(opt ClipBoardOption) *ClipBoard {
	c := &ClipBoard{
		opt: opt,
		pro: protoc{
			Prefix:   opt.Prefix,
			PageSize: opt.PageSize,
		},
	}
	return c
}

func (c *ClipBoard) Init() *ClipBoard {
	switch c.opt.Mode {
	case Client:
		go c.client().register(c).run()
	case Server:
		srv := c.server()
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
	for {
		select {
		case data := <-ch:
			if len(data) != 0 {
				log.Println("更新剪贴板数据:", string(data))
				go c.pub(data)
			}
		}
	}
}

func (c *ClipBoard) server() *ClipBoardServer {
	return &ClipBoardServer{
		Port: c.opt.Port,
		pro:  c.pro,
		cb:   c,
	}
}

func (c *ClipBoard) client() *ClipBoardClient {
	return &ClipBoardClient{
		ServerIP: c.opt.IP,
		Port:     c.opt.Port,
		pro:      c.pro,
	}
}

func clipboardWrite(body []byte) {
	log.Println("写入剪贴板s")
	ch := clipboard.Write(clipboard.FmtText, body)
	select {
	case _, ok := <-ch:
		if ok {
			log.Println("写入剪贴板成功")
		} else {
			log.Println("写入剪贴板失败")
		}
	}
}
