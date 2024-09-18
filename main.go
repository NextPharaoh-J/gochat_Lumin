package main

import (
	"flag"
	"fmt"
	"gochat_my/api"
	"gochat_my/connect"
	"gochat_my/logic"
	"gochat_my/site"
	"gochat_my/task"
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
		logic.New().Run()
	case "websocket":
		connect.New().Run()
	case "task":
		task.New().Run()
	case "api":
		api.New().Run()
	case "site":
		site.New().Run()
	default:
		fmt.Println("Exiting , module not support!")
		return
	}
	fmt.Printf("Run %s module done!\n", module)

	// 服务结束 退出提示
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit // 阻塞
	fmt.Println("Server exit")
}
