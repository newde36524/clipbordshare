package main

import (
	"fmt"

	"github.com/newde36524/clipboardshare"
	"github.com/spf13/cobra"
)

var (
	ip   string
	port int16
	mode string
)

func main() {
	config() //读取执行参数
	if mode == string(clipboardshare.Client) && (ip == "" || port == 0) {
		panic("客户端必须指明正确的服务器IP和端口,例子: go-shareclipbord -i 192.168.1.1 -p 9996 -m client")
	}
	if mode == string(clipboardshare.Server) && port == 0 {
		panic("服务端必须指明正确的端口,例子: go-shareclipbord -p 9996 -m server")
	}
	opt := clipboardshare.ClipBoardOption{
		Mode: clipboardshare.ClipBoardEnum(mode),
		IP:   ip,
		Port: port,
	}
	go clipboardshare.New(opt).Init().Run()
	fmt.Println("开始监听剪贴板")
	select {}
}

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
