package main

import (
	"hfproxy/pkg/config"
	"hfproxy/pkg/server"
)

func main() {
	// time.Sleep(30 * time.Second)
	// 载入并监听配置文件
	err := config.Init("hfproxy-config.yaml")
	if err != nil {
		panic(err)
	}
	// fmt.Printf("DB config = %v\n", config.GetConfig().DB)

	// err = mydb.Init()
	// if err != nil {
	// 	panic(err)
	// }

	// 初始化服务
	s := server.NewServer()
	s.Run()
}
