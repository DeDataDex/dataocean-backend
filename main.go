package main

import (
	_ "dataoceanbackend/routers"
	"flag"
	"fmt"

	beego "github.com/stonemeta/beego/server/web"
	"github.com/stonemeta/beego/server/web/filter/cors"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		remoteAddress := flag.String("chainApi", "127.0.0.1:1317", "rigister server in a application blockchain")
		dataDir := flag.String("fileDir", "./temp", "File Save Path")
		threshold := flag.String("threshold", "-200", "set judge threshold")

		key := flag.String("aesKey", "key_for_server_1", "set aeskey of every miner")

		flag.Parse()
		beego.AppConfig.Set("remoteAddress", *remoteAddress)
		beego.AppConfig.Set("aesKey", *key)
		fmt.Println("remoteAddress:", *remoteAddress)
		beego.AppConfig.Set("FileDir", *dataDir)
		beego.AppConfig.Set("threshold", *threshold)
	}

	cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	})

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		// 允许访问所有源
		AllowAllOrigins: true,
		// 可选参数"GET", "POST", "PUT", "DELETE", "OPTIONS" (*为所有)
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// 指的是允许的Header的种类
		AllowHeaders: []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		// 公开的HTTP标头列表
		ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		// 如果设置，则允许共享身份验证凭据，例如cookie
		AllowCredentials: true,
	}))

	beego.Run()
}
