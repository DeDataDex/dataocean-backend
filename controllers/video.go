package controllers

import (
	"cosmosVideoApi/models"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	beego "github.com/stonemeta/beego/server/web"
	"github.com/everFinance/goar/utils"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/golang-module/dongle"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"time"
)




type voucherInfo struct {
	address string
	level string
	sn string
	size string
}

// Operations about Users
type VideoController struct {
	beego.Controller
}

// @Title AddVideo
// @Description create users
// @Param	video		body 	models.Video	true		"body for user content"
// @Success 200 {int} models.User.Id
// @Failure 403 body is empty
// @router / [post]
func (v *VideoController) Post() {
	video := &models.Video{}
	v.ParseForm(video)

	//fmt.Println("remoteaddr:",u.Ctx.Request.Host)
	//fmt.Println(u.Ctx.Input.RequestBody)
	//if err:=json.Unmarshal(u.Ctx.Input.RequestBody, user); err != nil {
	//	fmt.Println("err:", err)
	//}
	log.Println("video:",video)
	vid := models.AddVideo(video)
	log.Println("vid:", vid)
	//将video信息转发给应用链
	//var msg string
	if succ, err := sendVideoInfoToChain(video); err != nil {
		v.Data["json"] = map[string]interface{}{"success": succ, "data":nil, "error":err}

	} else {
		v.Data["json"] = map[string]interface{}{"success": succ, "data":nil, "error":nil}
	}

	v.ServeJSON()
}


func (v *VideoController)GetVideo() {
	rr := v.Ctx.Request
	rw := v.Ctx.ResponseWriter
	var address,videoId,expire string
	//获取请求参数
	msg := v.GetString(":message")
	filename := v.GetString(":videoID")
	//对message进行Unescape
	msgUnescape, errUnescape := url.PathUnescape(msg)
	if errUnescape != nil {
		log.Println("errUnescape:",errUnescape)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	//解密message并获取参数
	decryptMsg := dongle.Decrypt.FromBase64String(msgUnescape).ByAes(cipher).ToString()
	address, videoId, expire = getDecryptMsg(decryptMsg)
	//校验参数是否满足条件
	expire1, errParseInt := strconv.ParseInt(expire, 10, 64)
	if  errParseInt != nil {
		log.Println("errParseInt:", errParseInt)
		sendErrorResponse(rw, models.ErrorInternalFaults)
		return
	}
	ext := strings.Split(filename,".")
	prefix := strings.Join(ext[:len(ext)-1],"")
	if videoId != prefix {
		log.Println("videoId与密文中的videoId不同")
		sendErrorResponse(rw,models.ErrorVideoIdError)
		return
	}
	if time.Now().Unix() > expire1 {
		log.Println("链接已过期")
		sendErrorResponse(rw, models.ErrorExpireError)
		return
	}
	dir, _ := beego.AppConfig.String("FileDir")
	thre, _ := beego.AppConfig.String("threshold")
	threshold, errThreshold := strconv.ParseInt(thre,10,64)
	if errThreshold != nil {
		log.Println("errThreshold:", errThreshold)
		sendErrorResponse(rw,models.ErrorInternalFaults)
		return
	}
	fmt.Println("threshold:", threshold)
	fmt.Println("dir:", dir)

	vl :=  dir + "/" + prefix + "/" + filename
	fmt.Println("vl:", vl)
	video, err := os.Open(vl)
	if err == errors.ErrNotFound {
		log.Println("该服务器没有此视频文件")
		sendErrorResponse(rw, models.ErrorFileError)
		return
	}
	defer video.Close()
	fileInfo, err :=  video.Stat()
	if err != nil {
		log.Println("Get FileInfo", err.Error())
		sendErrorResponse(rw,models.ErrorFileError)
		return
	}
	//beego.BConfig.WebConfig.ViewsPath
	var size int64

	filesize := fileInfo.Size()
	prexfile:=strings.Join(ext[:len(ext)-1], "")
	fmt.Println("prefile:", prexfile)
	key := []byte(address + prexfile)
	fmt.Println("key:", key)
	value, errGet := db.Get(key, nil)
	if errGet != errors.ErrNotFound {
		size, errSize := strconv.ParseInt(string(value),10,64)
		if errSize != nil {
			log.Println("parseInt:",errSize)
			sendErrorResponse(rw,models.ErrorInternalFaults)
			return
		}
		if ext[len(ext)-1] != "m3u8" {
			if size - filesize > threshold*1024*1024  {

				sendErrorResponse(rw, models.ErrorInsufficientBalance)
				return
			}

		}
	} else {
		size = 0
		if err := db.Put(key, []byte(strconv.FormatInt(size,10)),nil); err != nil {
			sendErrorResponse(rw,models.ErrorDBError)
		}
	}

	var start,end int64

	if rangeByte := rr.Header.Get("Range"); rangeByte != "" {
		fmt.Println("rangeByte:", rangeByte)
		if strings.Contains(rangeByte,"bytes=") && strings.Contains(rangeByte,"-") {
			fmt.Sscanf(rangeByte,"bytes=%d-%d", &start, &end)
			fmt.Println("start:",start)
			fmt.Println("end:",end)
			if end == 0 {
				end = fileInfo.Size() - 1
			}
			if start > end || start < 0 || end < 0 || end >= fileInfo.Size() {
				rw.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				log.Println("sendFile2 start:", start, "end:", end, "size:", fileInfo.Size())
				return
			}
			rw.Header().Add("Content-Length", strconv.FormatInt(end-start+1, 10))
			rw.Header().Add("Content-Range", fmt.Sprintf("bytes %v-%v/%v", start, end, fileInfo.Size()))
			rw.WriteHeader(http.StatusPartialContent)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			sendErrorResponse(rw, models.ErrorBadRequestError)
		}
	} else {
		rw.Header().Add("Content-Length", strconv.FormatInt(fileInfo.Size(),10))
		start = 0
		end = fileInfo.Size()-1
		fmt.Println("start:", start)
		fmt.Println("end:", end)
	}
	_, err = video.Seek(start,0)
	if err != nil {
		log.Println("file locate seek", err.Error())
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	//ext = strings.Split(fileInfo.Name(), ".")
	fmt.Println("ext:", ext[len(ext)-1])

	fmt.Println("content-type:", contentTypeMap[ext[len(ext)-1]])
	if ok := contentTypeMap[ext[len(ext)-1]]; ok != "" {
		rw.Header().Set("Content-Type", contentTypeMap[ext[len(ext)-1]])
	} else {
		rw.Header().Set("Content-Type", "application/octet-stream")
	}
	rw.Header().Add("Accept-Ranges", "bytes")

	//rw.Header().Add("Content-Disposition", "attachment; filename="+fileInfo.Name())
	rw.Header().Add("Content-Disposition", "attachment; filename="+fileInfo.Name())

	n := 512
	buf := make([]byte, n)
	for {
		if end-start+1 < int64(n) {
			n = int(end - start + 1)
			fmt.Println("n:", n)
		}
		_, err := video.Read(buf[:n])
		if err != nil {
			log.Println("1:", err)
			if err != io.EOF {
				log.Println("error:", err)
			}
			return
		}
		err = nil
		_, err = rw.Write(buf[:n])
		if err != nil {
			//log.Println(err, start, end, info.Size(), n)
			return
		}
		start += int64(n)
		if start >= end+1 {
			return
		}
	}
	size = size - filesize
	if errPut := db.Put(key, []byte(strconv.FormatInt(size,10)), nil); errPut != nil {
		log.Println("errPut:", errPut)
		sendErrorResponse(rw,models.ErrorDBError)
	}
	//v.Data["json"] = map[string]interface{}{"success": true, "msg": "文件传输成功"}
	//v.ServeJSON()
}



// @Title GetExtranetIp
// @Description get video by filename
// @Success 200 {object} models.Video
// @Failure 403  is empty
// @router /ip [get]
func (v *VideoController)GetIP() {
	responseClient, errClient := http.Get("http://myexternalip.com/raw") // 获取外网 IP
	if errClient != nil {
		fmt.Printf("获取外网 IP 失败，请检查网络\n")
		panic(errClient)
	}
	// 程序在使用完 response 后必须关闭 response 的主体。
	defer responseClient.Body.Close()

	body, _ := ioutil.ReadAll(responseClient.Body)
	clientIP := fmt.Sprintf("%s", string(body))
	fmt.Println(clientIP)

}

// @Title SendVoucher
// @Description send video voucher
// @Param	accountAddress 	query string	true
// @Param	level query	string	true
// @Param	sn query	string	true
// @Param	size  query	string	true
// @Success 200
// @Failure 403  is empty
// @router /sendVoucher [post]
func (v *VideoController)SendVoucher() {
	rw := v.Ctx.ResponseWriter
	var voucher models.VoucherInfo
	fmt.Println("voucher:", voucher)
	//获取签名参数
	if errParseForm := v.ParseForm(&voucher); errParseForm != nil {
		log.Println("errUnmarshal:", errParseForm)
		sendErrorResponse(rw, models.ErrorInternalFaults)
		return
	}
	//对签名参数进行解码
	newSize, errAtoi := strconv.Atoi(voucher.Size)
	fmt.Println("newSize:", newSize)
	if errAtoi != nil {
		log.Println("errAtoi:", errAtoi)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	var sizes int
	key := []byte(voucher.Account+voucher.Sn)
	data, errGet := db.Get(key,nil)
	fmt.Println("data:", data)
	if errGet == errors.ErrNotFound {
		sizes = 0
		fmt.Println("sizes:", sizes)
	} else {
		if errUnmarshal := json.Unmarshal(data, &sizes); errUnmarshal != nil {
			log.Println("errUnmarshal:", errUnmarshal)
			sendErrorResponse(rw, models.ErrorInternalFaults)
		}
		sizes = sizes + newSize
		fmt.Println("size:", sizes)
	}
	value, errMarshal := json.Marshal(sizes)
	fmt.Println("value:", value)
	if errMarshal != nil {
		log.Println("errMarshal:", errMarshal)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	if errPut :=db.Put(key,value,nil); errPut != nil {
		log.Println("errPut:", errPut)
		sendErrorResponse(rw, models.ErrorDBError)
	}
	var msg models.SettleInfo
	msg.UserAddress = voucher.Account
	msg.TimeStamp = time.Now().Format("2006/01/02 15:04:05")
	msg.Charge = voucher.Size
	if err := mQueue.Publish("sendTx", msg); err != nil {
		log.Println("mq publish error:", err.Error())
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}

	v.Data["json"] = map[string]interface{}{"result": true, "msg": "签名校验成功" }
	v.ServeJSON()

}

// @Title settlement
// @Description chain token settle
// @Param		videoID  path	string	true
// @Success 200
// @Failure 403  is empty
// @router /settle [post]
func (v *VideoController)Settlement() {
	var settle *models.SettleInfo
	videoID := v.GetString("videoID")
	fmt.Println("videoID:", videoID)
	if err := v.ParseForm(settle); err != nil {
		log.Println("settle parse form:", err)
		return
	}

	v.ServeJSON()
}

// @Title getVideoFromAr
// @Description getVideoFromAr
// @Param		txUrl path string	true
// @Success 200
// @Failure 403  is empty
// @router /getVideoFromAr [get]
func (v *VideoController)GetVideoFromAr() {
	txId := v.GetString("txUrl")
	fmt.Println("txID:", txId)
	arNode := "http://localhost:1984"

	c := goar.NewClient(arNode)
	transaction, errByID := c.GetTransactionByID(txId)
	if errByID != nil {
		log.Println("get transaction:", errByID)
	}
	fmt.Println("transaction:", transaction)
	//data,errData := c.GetTransactionData(txId)
	//if errData != nil {
	//	log.Println("get transaction data:", errData)
	//	return
	//}
}

// @Title sendVideoToAr
// @Description sendVideoToAr
// @Success 200
// @Failure 403  is empty
// @router /sendVideoToAr [post]
func (v *VideoController)SendVideoToAr() {
	arNode := "http://localhost:1984"
	w, err := goar.NewWalletFromPath("./conf/account2.json", arNode)
	//fmt.Println("owner:",w.Signer.Owner())
	if err != nil {
		log.Println("create new wallet:", err)
		return
	}
	//fmt.Println("w:", w)
	data, errReadFile := ioutil.ReadFile("./models/1/huahai2.ts")
	//fmt.Println("data:",data)
	if errReadFile != nil {
		log.Println("read file:", errReadFile)
		return
	}
	tags := []types.Tag{{Name: "Content-Type", Value: "video/mp4"}, {Name: "goar", Value: "testdata"}}
	tx, errSendData := assemblyDataTx(data,w,tags)
	if errSendData != nil {
		log.Println("send data speedup:", errSendData)
		return
	}
	fmt.Println("tx:", tx.ID)
	v.Data["json"] = map[string]interface{}{"success": true}
	v.ServeJSON()

}

func assemblyDataTx(bigData []byte, wallet *goar.Wallet, tags []types.Tag) (*types.Transaction, error) {
	reward, err := wallet.Client.GetTransactionPrice(bigData, nil)
	if err != nil {
		return nil, err
	}
	tx := &types.Transaction{
		Format:   2,
		Target:   "",
		Quantity: "0",
		Tags:     utils.TagsEncode(tags),
		Data:     utils.Base64Encode(bigData),
		DataSize: fmt.Sprintf("%d", len(bigData)),
		Reward:   fmt.Sprintf("%d", reward),
	}
	anchor, err := wallet.Client.GetTransactionAnchor()
	if err != nil {
		return nil, err
	}
	tx.LastTx = anchor
	tx.Owner = wallet.Owner()

	signData, err := utils.GetSignatureData(tx)
	if err != nil {
		return nil, err
	}

	sign, err := wallet.Signer.SignMsg(signData)
	if err != nil {
		return nil, err
	}

	txHash := sha256.Sum256(sign)
	tx.ID = utils.Base64Encode(txHash[:])

	tx.Signature = utils.Base64Encode(sign)
	return tx, nil
}


func sendVideoInfoToChain(video *models.Video) (string, error) {
	for {
		videoinfo,_ :=json.Marshal(video)
		remoteAddress, _ := beego.AppConfig.String("remoteAddress")
		sendChainApi, _ := beego.AppConfig.String("chainApi")

		req, err := http.NewRequest("POST", fmt.Sprintf(sendChainApi, remoteAddress), strings.NewReader(string(videoinfo)))
		if err != nil {
			fmt.Print("Register: http.NewRequest ", err.Error())
			return "false",err
		}


		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			return "false",err
		}
		//var respBody string
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Println("同步video信息成功:"+string(respBody))
		//fmt.Println(respBody)

		resp.Body.Close()

		return "true",nil

	}
}

func getDecryptMsg(src string) (address string,videoId string,expire string) {
	strs := strings.Split(src, ",")
	address = strings.Split(strs[0],"=")[1]
	videoId = strings.Split(strs[1], "=")[1]
	expire = strings.Split(strs[2], "=")[1]
	return
}








