package controllers

import (
	// "dataoceanbackend/types"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"dataoceanbackend/models"
	"dataoceanbackend/mq"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/go-resty/resty/v2"
	beego "github.com/stonemeta/beego/server/web"
	// "github.com/cosmos/cosmos-sdk/codec"
	// codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	// "github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/golang-module/dongle"
	"github.com/syndtr/goleveldb/leveldb"
)

var account string
var contentTypeMap = map[string]string{
	// "mp4": "video/mp4",
	"asf":   "video/x-ms-asf",
	"asx":   "video/x-ms-asf",
	"avi":   "video/avi",
	"m1v":   "video/x-mpeg",
	"m2v":   "video/x-mpeg",
	"m3u":   "audio/mpegurl",
	"m4e":   "video/mpeg4",
	"movie": "video/x-sgi-movie",
	"mp2v":  "video/mpeg",
	"mp4":   "video/mp4",
	"mpa":   "video/x-mpg",
	"mpe":   "video/x-mpeg",
	"mpeg":  "video/mpg",
	"mpg":   "video/mpg",
	"mps":   "video/x-mpeg",
	"mpv":   "video/mpg",
	"mpv2":  "video/mpeg",
	"wmv":   "video/x-ms-wmv",
	"wmx":   "video/x-ms-wmx",
	"ts":    "video/vnd.dlna.mpeg-tts",
}
var mQueue *mq.Client
var db *leveldb.DB
var cipher *dongle.Cipher
var ctx sdk.Context
var txBuilder client.TxBuilder

func init() {
	var err error
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
		ch, err := m.Subscribe(top)
		if err != nil {
			fmt.Printf("sub top:%s failed\n", top)
		}
		for {
			select {
			case <-t.C:
				val := m.GetPayLoad(ch)
				if val != nil {
					go sendSettleMsg(ctx, val.([]byte))
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

func sendSettleMsg(ctx sdk.Context, val []byte) {
	remoteAddress, errAppConfig := beego.AppConfig.String("remoteAddress")
	if errAppConfig != nil {
		log.Println("errAppConfigGet:", errAppConfig)
		return
	}
	client := resty.New()

	fmt.Println("sendSettleMsg:", string(val))
	var i int
	for i = 0; i < 5; i++ {
		result := struct {
			TxResponse struct {
				Code   int    `json:"code"`
				RawLog string `json:"raw_log"`
			} `json:"tx_response"`
		}{}
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]string{
				"tx_bytes": string(val),
				"mode":     "BROADCAST_MODE_BLOCK",
			}).
			SetResult(&result).
			Post(fmt.Sprintf("http://%s/cosmos/tx/v1beta1/txs", remoteAddress))
		if err != nil {
			fmt.Println("sendSettleMsg err", err.Error())
			time.Sleep(time.Second)
			continue
		}
		fmt.Printf("sendSettleMsg ret: code=%d rawlog=%s resp=%s", result.TxResponse.Code, result.TxResponse.RawLog, resp.String())

		break
	}
	if i >= 5 {
		log.Println("not settleMsg success", string(val))
	}
}
