package main

import (
	"github.com/kanzx/shuidiVPN"
	"log"
)

var (
	version = "测试版"
)

func main()  {
	// 读取默认配置
	config := &shuidiVPN.Config{}
	config.ReadConfig()

	listenServer, err := shuidiVPN.NewServer(config.ServerAddr, config.Password)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("水滴VPN-Server:%s启动成功，监听端口 %s", version, config.ServerAddr)

	err = listenServer.Listen()
	if err != nil {
		log.Fatalln(err)
	}
}
