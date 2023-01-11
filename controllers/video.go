package controllers

import (
	"crypto/sha256"
	"dataoceanbackend/models"
	"encoding/base64"
	"encoding/json"
	"fmt"
	clientTx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	beego "github.com/stonemeta/beego/server/web"
	//tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	cosTypes "dataoceanbackend/types"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/everFinance/goar"
	"github.com/everFinance/goar/types"
	"github.com/everFinance/goar/utils"
	"github.com/golang-module/dongle"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type voucherInfo struct {
	address string
	level   string
	sn      string
	size    string
}

// Operations about Users
type VideoController struct {
	beego.Controller
}

func (v *VideoController) GetVideo() {

	rr := v.Ctx.Request
	rw := v.Ctx.ResponseWriter
	var address, videoId, expire string
	//获取请求参数
	msg := v.GetString(":message")
	filename := v.GetString(":videoID")
	//对message进行Unescape
	msgUnescape, errUnescape := url.PathUnescape(msg)
	if errUnescape != nil {
		log.Println("errUnescape:", errUnescape)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	fmt.Println("msg:", msg)
	fmt.Println("mgunescape:", msgUnescape)
	//解密message并获取参数
	//解密message并获取参数
	aeskeying, _ := beego.AppConfig.String("aesKey")
	fmt.Println("key:", aeskeying)
	cipher = dongle.NewCipher()
	cipher.SetMode(dongle.ECB)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(aeskeying)
	decryptMsg := dongle.Decrypt.FromBase64String(msgUnescape).ByAes(cipher).ToString()
	fmt.Println("decryptmsg:", decryptMsg)
	address, videoId, expire = getDecryptMsg(decryptMsg)
	//校验参数是否满足条件
	expire1, errParseInt := strconv.ParseInt(expire, 10, 64)
	if errParseInt != nil {
		log.Println("errParseInt:", errParseInt)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	ext := strings.Split(filename, ".")
	prefix := strings.Join(ext[:len(ext)-1], "")

	if ext[len(ext)-1] != "m3u8" {
		temp := strings.Split(prefix, "-")
		prefix = strings.Join(temp[:len(temp)-1], "")
	}

	if videoId != prefix {
		log.Println("videoId与密文中的videoId不同")
		sendErrorResponse(rw, models.ErrorVideoIdError)

	}
	fmt.Println("expire1:", expire1)
	if time.Now().Unix() > expire1 {
		log.Println("链接已过期")
		sendErrorResponse(rw, models.ErrorExpireError)
		return
	}
	dir, _ := beego.AppConfig.String("FileDir")
	thre, _ := beego.AppConfig.String("threshold")
	threshold, errThreshold := strconv.ParseInt(thre, 10, 64)
	if errThreshold != nil {
		log.Println("errThreshold:", errThreshold)
		sendErrorResponse(rw, models.ErrorInternalFaults)
		return
	}
	fmt.Println("threshold:", threshold)
	fmt.Println("dir:", dir)

	vl := dir + "/" + prefix + "/" + filename
	fmt.Println("vl:", vl)
	video, err := os.Open(vl)
	if err == errors.ErrNotFound {
		log.Println("没有找到视频")
		sendErrorResponse(rw, models.ErrorFileError)
		return
	}
	defer video.Close()
	fileInfo, err := video.Stat()
	if err != nil {
		log.Println("Get FileInfo", err.Error())
		sendErrorResponse(rw, models.ErrorFileError)
		return
	}
	//beego.BConfig.WebConfig.ViewsPath
	var size int64

	filesize := fileInfo.Size()
	prexfile := strings.Join(ext[:len(ext)-1], "")
	fmt.Println("prefile:", prexfile)
	key := []byte(address + prexfile)
	fmt.Println("key:", key)
	value, errGet := db.Get(key, nil)
	if errGet != errors.ErrNotFound {
		size, errSize := strconv.ParseInt(string(value), 10, 64)
		if errSize != nil {
			log.Println("parseInt:", errSize)
			sendErrorResponse(rw, models.ErrorInternalFaults)
			return
		}
		if ext[len(ext)-1] != "m3u8" {

			if filesize-size > threshold*1024*1024 {

				sendErrorResponse(rw, models.ErrorInsufficientBalance)
				return
			}

		}
	} else {
		size = 0
		if err := db.Put(key, []byte(strconv.FormatInt(size, 10)), nil); err != nil {
			sendErrorResponse(rw, models.ErrorDBError)
		}
	}

	var start, end int64

	if rangeByte := rr.Header.Get("Range"); rangeByte != "" {
		fmt.Println("rangeByte:", rangeByte)
		if strings.Contains(rangeByte, "bytes=") && strings.Contains(rangeByte, "-") {
			fmt.Sscanf(rangeByte, "bytes=%d-%d", &start, &end)
			fmt.Println("start:", start)
			fmt.Println("end:", end)
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
		rw.Header().Add("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
		start = 0
		end = fileInfo.Size() - 1
		fmt.Println("start:", start)
		fmt.Println("end:", end)
	}
	_, err = video.Seek(start, 0)
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
	if errPut := db.Put(key, []byte(strconv.FormatInt(size, 10)), nil); errPut != nil {
		log.Println("errPut:", errPut)
		sendErrorResponse(rw, models.ErrorDBError)
	}
}

// @Title GetExtranetIp
// @Description get video by filename
// @Success 200 {object} models.Video
// @Failure 403  is empty
// @router /ip [get]
func (v *VideoController) GetIP() {
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

func (v *VideoController) SendVoucher() {

	rw := v.Ctx.ResponseWriter

	//获取签名参数
	paySign := v.GetString("paySign")
	fmt.Println("paySign:", paySign)
	payData := v.GetString("payData")
	fmt.Println("payData:", payData)
	voucherSign, errDecode := parseVoucherSign(paySign)
	if errDecode != nil {
		log.Println("errDecode:", errDecode)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}

	//对签名参数进行解码
	voucherData, errParsePayData := parsePayData(payData, voucherSign.PayPublickey)
	if errParsePayData != nil {
		log.Println("errParsePayData", errParsePayData)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	newSize := voucherData.ReceivedSizeMB
	var sizes uint64
	key := []byte(voucherSign.Creator + strconv.FormatUint(voucherSign.VidoId, 10))
	data, errGet := db.Get(key, nil)

	if errGet == errors.ErrNotFound {
		sizes = 0
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
	if errPut := db.Put(key, value, nil); errPut != nil {
		log.Println("errPut:", errPut)
		sendErrorResponse(rw, models.ErrorDBError)
	}
	var msg []byte
	msg, errSubmit := makeSubmitPaysign(voucherSign.Creator, paySign, payData)
	if errSubmit != nil {
		log.Println("errSumit:", errSubmit)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}

	if err := mQueue.Publish("sendTx", msg); err != nil {
		log.Println("mq publish error:", err.Error())
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}

	sendNormalResponse(rw, "签名校验成功", 201)

}

func (v *VideoController) Settlement() {
	var settle *models.SettleInfo
	videoID := v.GetString("videoID")
	fmt.Println("videoID:", videoID)
	if err := v.ParseForm(settle); err != nil {
		log.Println("settle parse form:", err)
		return
	}

	v.ServeJSON()
}

func (v *VideoController) GetVideoFromAr() {
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

func (v *VideoController) SendVideoToAr() {
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
	tx, errSendData := assemblyDataTx(data, w, tags)
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
		videoinfo, _ := json.Marshal(video)
		remoteAddress, _ := beego.AppConfig.String("remoteAddress")
		sendChainApi, _ := beego.AppConfig.String("chainApi")

		req, err := http.NewRequest("POST", fmt.Sprintf(sendChainApi, remoteAddress), strings.NewReader(string(videoinfo)))
		if err != nil {
			fmt.Print("Register: http.NewRequest ", err.Error())
			return "false", err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			return "false", err
		}
		//var respBody string
		respBody, _ := ioutil.ReadAll(resp.Body)
		log.Println("同步video信息成功:" + string(respBody))
		//fmt.Println(respBody)

		resp.Body.Close()

		return "true", nil

	}
}

//func parseMsg(src string) models.VoucherInfo {
//	var voucherInfo models.VoucherInfo
//
//	strs := strings.Split(src, " ")
//	creator := strings.Split(strs[0], ":")[1]
//	videoid := strings.Split(strs[1], ":")[1]
//	level := strings.Split(strs[2], ":")[1]
//	sn := strings.Split(strs[3], ":")[1]
//	receiveMB := strings.Split(strs[4], ":")[1]
//	time := strings.Split(strs[5], ":")[1]
//	voucherInfo.Sn, _ = strconv.ParseUint(sn, 10, 64)
//	voucherInfo.Level, _ = strconv.ParseUint(level, 10, 64)
//	voucherInfo.VidoId, _ = strconv.ParseUint(videoid, 10, 64)
//	voucherInfo.ReceivedSizeMB, _ = strconv.ParseUint(receiveMB, 10, 64)
//	voucherInfo.Timestamp, _ = strconv.ParseUint(time, 10, 64)
//	voucherInfo.Creator = creator
//
//	return voucherInfo
//}

func getDecryptMsg(src string) (address string, videoId string, expire string) {
	strs := strings.Split(src, ",")
	address = strings.Split(strs[0], "=")[1]
	videoId = strings.Split(strs[1], "=")[1]
	expire = strings.Split(strs[2], "=")[1]
	return
}

func makeSubmitPaysign(creator string, paySign string, payData string) ([]byte, error) {
	//var priv secp256k1.PrivKey
	var accountNumber uint64
	var sequence uint64
	//priv, err := getPrivKey(creator)
	priv := secp256k1.GenPrivKey()
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return []byte{}, err
	//}
	pub := priv.PubKey()
	addr := sdk.AccAddress(pub.Address())
	fmt.Println("pub:", pub)
	fmt.Println("addr:", addr.String())

	accountNumber, sequence, errGetAccount := getAccountNumSequence(addr.String())
	if errGetAccount != nil {
		log.Println("errGetAccount:", errGetAccount)
		return []byte{}, errGetAccount
	}

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &cosTypes.MsgSubmitPaySign{})
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(protoCodec, tx.DefaultSignModes)
	txBuilder = txConfig.NewTxBuilder()

	msg1 := cosTypes.NewMsgSubmitPaySign(addr.String(), paySign, payData)
	errSet := txBuilder.SetMsgs(msg1)
	if errSet != nil {
		fmt.Println(errSet)
		return []byte{}, errSet
	}
	txJSONBytes, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	fmt.Println(string(txJSONBytes))

	sigV2 := signing.SignatureV2{
		PubKey: pub,
		Data: &signing.SingleSignatureData{
			SignMode:  txConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: 0,
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	signerData := xauthsigning.SignerData{
		ChainID:       "dataocean",
		AccountNumber: accountNumber,
		Sequence:      sequence,
	}

	sigV2, err = clientTx.SignWithPrivKey(
		txConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, txConfig, sequence)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	txJSONBytes, err = txConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	fmt.Println(string(txJSONBytes))

	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	txBytesBase64 := base64.StdEncoding.EncodeToString(txBytes)
	fmt.Println(string(txBytesBase64))

	return txBytes, nil

}

func getAccountNumSequence(address string) (uint64, uint64, error) {
	remoteAddress, err := beego.AppConfig.String("remoteAddress")
	if err != nil {
		log.Println("getremoteAddress:", remoteAddress)
		return 0, 0, err
	}
	req, errRequest := http.NewRequest("POST", fmt.Sprintf("http://%s/cosmos/auth/v1beta1/accounts/%s", remoteAddress, address), strings.NewReader(""))
	if errRequest != nil {
		log.Println("errRequest:", errRequest)
		return 0, 0, errRequest
	}
	res, errRes := http.DefaultClient.Do(req)
	if errRes != nil {
		log.Println("errRes:", errRes)
		return 0, 0, errRequest
	}
	respBody, err := ioutil.ReadAll(res.Body)
	num, sequence, errParse := parseResNum(string(respBody))
	if errParse != nil {
		log.Println("errParse:", errParse)
		return 0, 0, errParse
	}
	return num, sequence, nil

}

func parseResNum(res string) (uint64, uint64, error) {
	var result models.GetAccountRes
	if err := json.Unmarshal([]byte(res), &result); err != nil {
		log.Println("err:", err)
		return 0, 0, err
	}
	num, seq := result.AccountNum, result.Squence
	return num, seq, nil

}

func getCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

//func getPrivKey(addr string) (cryptotypes.PrivKey, error) {
//	acc, err := sdk.AccAddressFromBech32(addr)
//	if err != nil {
//		return nil, err
//	}
//
//	kr, err := keyring.New("dataocean", keyring.BackendTest, "~/.dataocean", nil, getCodec())
//	k, err := kr.KeyByAddress(acc)
//	if err != nil {
//		return nil, err
//	}
//	rl := k.GetLocal()
//	privKey := rl.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
//	return privKey, nil
//}

func parsePayData(payData string, publicKey string) (*models.VoucherPayData, error) {
	var voucherData *models.VoucherPayData

	decrptData := dongle.Decrypt.FromHexString(payData).ByRsa(publicKey).ToString()
	fmt.Println("decrptData:", decrptData)
	if err := json.Unmarshal([]byte(decrptData), voucherData); err != nil {
		log.Println("err:", err)
		return nil, err
	}

	return voucherData, nil
}

func parseVoucherSign(paySign string) (*models.VoucherPaySign, error) {
	var voucherSign *models.VoucherPaySign
	parseSign, err := parsePaySign(sdk.Context{}, paySign)
	if err != nil {
		log.Println("parsePaySign:", err)
		return nil, err
	}
	voucherSign.PayPublickey = parseSign.PayPublicKey
	voucherSign.VidoId = parseSign.VideoId
	voucherSign.Creator = parseSign.Creator

	return voucherSign, nil

}

func parsePaySign(ctx2 sdk.Context, paySignStr string) (*cosTypes.MsgPaySign, error) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &cosTypes.MsgPaySign{})
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(protoCodec, tx.DefaultSignModes)

	txBytes, err := base64.StdEncoding.DecodeString(paySignStr)
	if err != nil {
		return nil, err
	}
	fmt.Println("txBytes:", txBytes)
	theTx, err := txConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, err
	}

	msgs := theTx.GetMsgs()
	if len(msgs) == 0 {
		return nil, errors.New("signature message is empty")
	}
	return msgs[0].(*cosTypes.MsgPaySign), nil
}
