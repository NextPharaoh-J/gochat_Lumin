package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gochat_my/api/router"
	"gochat_my/api/rpc"
	"gochat_my/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Chat struct{}

func New() *Chat {
	return &Chat{}
}

func (c *Chat) Run() {
	rpc.InitLogicRpcClient()
	r := router.Register()
	runMode := config.GetGinRunMode()
	logrus.Infof("server start , now is running on %s", runMode)
	gin.SetMode(runMode)
	apiConfig := config.Conf.Api
	port := apiConfig.ApiBase.ListenPort
	//	flag.Parse()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Errorf("start listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	logrus.Infof("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server Shutdown Failed:%+v", err)
	}
	logrus.Infof("Server exiting")
	os.Exit(0)
}
