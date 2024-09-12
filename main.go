package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// 模式获取 -module [logic , websocket , task , api , site]
	var module string
	flag.StringVar(&module, "module", "", "module name")
	flag.Parse()
	fmt.Printf("Start run %s module\n", module)

	// 启动服务
	switch module {
	case "logic":

		return
	case "websocket":
		return
	case "task":
		return
	case "api":
		return
	case "site":
		//return
	default:
		fmt.Println("Exiting , module not support!")
		//return
	}
	fmt.Printf("Run %s module done!\n", module)

	// 服务结束 退出提示
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit // 阻塞
	fmt.Println("Server exit")
}
