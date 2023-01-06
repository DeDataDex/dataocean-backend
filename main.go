package main

import (
	_ "cosmosVideoApi/routers"
	"flag"
	"fmt"
	beego "github.com/stonemeta/beego/server/web"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {


	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		remoteAddress := flag.String("chainApi", "127.0.0.1:16800", "rigister server in a application blockchain")
		dataDir := flag.String("fileDir", "./temp", "File Save Path")
		threshold := flag.String("threshold", "-200", "set judge threshold")
		flag.Parse()
		beego.AppConfig.Set("remoteAddress", *remoteAddress)
		//beego.AppConfig.Set("chainApi", )
		fmt.Println("remoteAddress:", *remoteAddress)
		beego.AppConfig.Set("FileDir", *dataDir)
		beego.AppConfig.Set("threshold", *threshold)
	}

	//localAddress := flag.String("local", "127.0.0.1:26800", "c2 host http listen address")
	go func() {
		//Register(*remoteAddress)
		//http.Handle("/video", )
		http.Handle("/", http.FileServer(http.Dir("./temp")))
		http.ListenAndServe(":8123", nil)
	}()
	beego.Run()
}

func Register(remoteAddress string) {
	for {
		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/proxy/register", remoteAddress), strings.NewReader("rigister a server"))
		if err != nil {
			fmt.Print("Register: http.NewRequest ", err.Error())
		}


		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("rigister ok!")
		}
		//var respBody string
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Println("收到应用链的注册回复:"+string(respBody))
		//fmt.Println(respBody)

		resp.Body.Close()


		time.Sleep(30 * time.Second)
	}
}
