package main

import (
	_ "dataoceanbackend/routers"
	"flag"
	"fmt"
	beego "github.com/stonemeta/beego/server/web"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		remoteAddress := flag.String("chainApi", "127.0.0.1:1317", "rigister server in a application blockchain")
		dataDir := flag.String("fileDir", "./temp", "File Save Path")
		threshold := flag.String("threshold", "200", "set judge threshold")

		key := flag.String("aesKey", "key_for_server_1", "set aeskey of every miner")

		flag.Parse()
		beego.AppConfig.Set("remoteAddress", *remoteAddress)
		beego.AppConfig.Set("aesKey", *key)
		fmt.Println("remoteAddress:", *remoteAddress)
		beego.AppConfig.Set("FileDir", *dataDir)
		beego.AppConfig.Set("threshold", *threshold)
	}

	beego.Run()
}
