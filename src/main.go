package main

import (
	"fmt"
	"matrixChannel/config"
	"matrixChannel/handler"
	"matrixChannel/util"
	"os"
	"os/signal"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	conf := util.GetConfig()

	// 用于 OAuth 验证
	go HttpSvr(conf.Service)
	// 后台数据同步
	go BackSvr(conf)
	<-c
}

// BackSvr 后台数据同步
func BackSvr(conf *config.Config) {
	timeTickerChan := time.Tick(conf.Service.RefreshInterval)
	for {
		if len(conf.User) < 1 {
			fmt.Println("未配置用户， end.")
			return
		}
		for _, userConf := range conf.User {
			run(conf.Service, userConf)
		}
		<-timeTickerChan
	}
}

func run(serviceConf *config.ServiceConfig, userConf *config.UserConfig) {
	fmt.Printf("User[%s] work begin\n", userConf.TapdOwner)
	start := time.Now()
	for _, v := range userConf.Enable {
		if v == "story" {
			_ = handler.TapdStoryProvider.New(serviceConf, userConf).Do()
		}
		if v == "task" {
			_ = handler.TapdTaskProvider.New(serviceConf, userConf).Do()
		}
	}
	finish := time.Now()
	fmt.Printf("User[%s] work end. Time cost [%f]s. Wait [%d]s for next order.\n", userConf.TapdOwner, finish.Sub(start).Seconds(), int(serviceConf.RefreshInterval.Seconds())-finish.Second())
}
