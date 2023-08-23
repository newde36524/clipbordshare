package clipboardshare

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

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
	opt           ClipBoardOption
	pub           func([]byte)
	isReceiveData bool
	lock          sync.RWMutex
	cond          *sync.Cond
}

func New(opt ClipBoardOption) *ClipBoard {
	c := &ClipBoard{
		opt: opt,
	}
	c.cond = sync.NewCond(&c.lock)
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
	temp := int32(0)
	for data := range ch {
		if len(data) != 0 {
			c.lock.Lock()
			for atomic.LoadInt32(&temp) == 0 {
				c.cond.Wait()
			}
			atomic.StoreInt32(&temp, 0)

			log.Println("更新剪贴板数据:", string(data))
			c.pub(data)

			c.lock.Unlock()
			c.cond.Signal()
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
	<-clipboard.Write(clipboard.FmtText, body)
}
