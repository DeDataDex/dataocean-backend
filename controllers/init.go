package controllers

import (
	"cosmosVideoApi/models"
	"cosmosVideoApi/mq"
	"cosmossdk.io/simapp"
	"encoding/json"
	"fmt"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/types/tx"
	beego "github.com/stonemeta/beego/server/web"
	"github.com/syndtr/goleveldb/leveldb"
	logging "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc"
	"os"
	//"github.com/cosmos/cosmos-sdk/types/tx/signing"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	//"github.com/golang-module/dongle"
	"github.com/golang-module/dongle"
	"io"
	//"io/ioutil"
	"log"
	"net/http"
	//"strings"
	"time"
)

var account string
var contentTypeMap = map[string]string{
	//"mp4": "video/mp4",
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
var appCos *simapp.SimApp
var cipher *dongle.Cipher
var ctx sdk.Context

func init() {
	var err error
	cosmosDb := dbm.NewMemDB()
	logger := logging.NewTMLogger(logging.NewSyncWriter(os.Stdout))
	appCos = simapp.NewSimApp(logger, cosmosDb, nil, true, simtestutil.NewAppOptionsWithFlagHome(simapp.DefaultNodeHome))
	ctx = appCos.NewContext(true, tmproto.Header{Height: appCos.LastBlockHeight()})
	key, _ := beego.AppConfig.String("aesKey")
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
		ch, err := m.Subscribe(top)
		if err != nil {
			fmt.Printf("sub top:%s failed\n", top)
		}
		for {
			select {
			case <-t.C:
				val := m.GetPayLoad(ch)
				if val != nil {
					go sendSettleMsg(ctx, val)
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

func sendSettleMsg(ctx sdk.Context, val interface{}) {
	remoteAddress, errAppConfig := beego.AppConfig.String("remoteAddress")
	if errAppConfig != nil {
		log.Println("errAppConfigGet:", errAppConfig)
		return
	}
	fmt.Println("val:", val)
	for {
		grpcConn, errDial := grpc.Dial(remoteAddress, grpc.WithInsecure())
		if errDial != nil {
			log.Println("errDail:", errDial)
		}
		defer grpcConn.Close()
		txclient := tx.NewServiceClient(grpcConn)
		grpcRes, err := txclient.BroadcastTx(ctx,
			&tx.BroadcastTxRequest{
				Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
				TxBytes: val.([]byte)},
		)
		if err != nil {
			log.Println("err:", err)
		}
		if grpcRes.TxResponse.Code == 0 {
			break
		}
		time.Sleep(time.Second)
	}
}
