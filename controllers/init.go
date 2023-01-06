package controllers

import (
	"cosmosVideoApi/models"
	"cosmosVideoApi/mq"
	"encoding/json"
	"fmt"
	beego "github.com/stonemeta/beego/server/web"
	"github.com/syndtr/goleveldb/leveldb"
	//"github.com/golang-module/dongle"
	"github.com/golang-module/dongle"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)
var sendChainApi = "http://%s/proxy/register"
var contentTypeMap = map[string]string{
	//"mp4": "video/mp4",
	"asf": "video/x-ms-asf",
	"asx": "video/x-ms-asf",
	"avi": "video/avi",
	"m1v": "video/x-mpeg",
	"m2v": "video/x-mpeg",
	"m3u": "audio/mpegurl",
	"m4e": "video/mpeg4",
	"movie": "video/x-sgi-movie",
	"mp2v": "video/mpeg",
	"mp4": "video/mp4",
	"mpa": "video/x-mpg",
	"mpe": "video/x-mpeg",
	"mpeg": "video/mpg",
	"mpg": "video/mpg",
	"mps": "video/x-mpeg",
	"mpv": "video/mpg",
	"mpv2": "video/mpeg",
	"wmv": "video/x-ms-wmv",
	"wmx": "video/x-ms-wmx",
	"ts": "video/vnd.dlna.mpeg-tts",
}
var mQueue *mq.Client
var db *leveldb.DB

var cipher *dongle.Cipher
var key string
func init() {
	var err error
	key = "key_for_server_2"
	cipher = dongle.NewCipher()
	cipher.SetMode(dongle.ECB)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(key)
	db, err = leveldb.OpenFile("./path/db", nil)
	if err != nil {
		log.Println("level db open file:", err)
		panic(err)
	}
	mQueue = mq.NewClient()
	mQueue.SetConditions(1000)
	go func(m *mq.Client, top string) {
		var chl chan int
		log.Println("初始化消息队列")
		t := time.NewTicker(time.Second)
		defer t.Stop()
		ch,err := m.Subscribe(top)
		if err != nil{
			fmt.Printf("sub top:%s failed\n",top)
		}
		for {
			select {
			case <-t.C:
				val := m.GetPayLoad(ch)
				if val != nil {
					go sendSettleMsg(val)
				}
			default:
			}
		}
		<-chl

	}(mQueue, "sendTx")


}

func sendErrorResponse(w http.ResponseWriter, errResp models.ErrResponse) {
	w.WriteHeader(errResp.HttpSC)

	resStr, _ := json.Marshal(&errResp.Error)
	io.WriteString(w, string(resStr))
}

func sendNormalResponse(w http.ResponseWriter, resp string, sc int) {
	w.WriteHeader(sc)
	io.WriteString(w, resp)
}

func sendSettleMsg(val interface{}) {
	var settleResponse models.SettleResponse
	remoteAddress,errAppConfig := beego.AppConfig.String("remoteAddress")
	if errAppConfig != nil {
		log.Println("errAppConfigGet:", errAppConfig)
		return
	}
	fmt.Println("val:", val)
	for {
		req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/proxy/register", remoteAddress), strings.NewReader("rigister a server"))

		if err != nil {
			log.Fatalf("http.NewRequest error: ", err.Error())
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("receive response error:", err.Error())
		} else {
			respBody, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("respBody:", string(respBody))
			//if err := json.Unmarshal(respBody, settleResponse); err != nil {
			//	log.Println("unmarshal error:", err.Error())
			//}
			settleResponse.Result = "success"
			if settleResponse.Result == "success" {
				break
			}
			time.Sleep(time.Second)
		}

	}


}

