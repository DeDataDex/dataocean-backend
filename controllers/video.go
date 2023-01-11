package controllers

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"dataoceanbackend/models"
	cosTypes "dataoceanbackend/types"
	clientTx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	beego "github.com/stonemeta/beego/server/web"

	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/golang-module/dongle"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

// Operations about Users
type VideoController struct {
	beego.Controller
}

func (v *VideoController) GetVideo() {
	var mutexDB sync.Mutex
	rr := v.Ctx.Request
	rw := v.Ctx.ResponseWriter
	var address, videoId, expire string
	// 获取请求参数
	msg := v.GetString(":message")
	filename := v.GetString(":videoID")
	// 对message进行Unescape
	msgUnescape, errUnescape := url.PathUnescape(msg)
	if errUnescape != nil {
		fmt.Println("errUnescape:", errUnescape)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	// 解密message并获取参数
	// 解密message并获取参数
	aeskeying, _ := beego.AppConfig.String("aesKey")
	fmt.Println("key:", aeskeying)
	cipher = dongle.NewCipher()
	cipher.SetMode(dongle.ECB)
	cipher.SetPadding(dongle.PKCS7)
	cipher.SetKey(aeskeying)
	decryptMsg := dongle.Decrypt.FromBase64String(msgUnescape).ByAes(cipher).ToString()
	fmt.Println("decryptmsg:", decryptMsg)
	address, videoId, expire = getDecryptMsg(decryptMsg)
	// 校验参数是否满足条件
	expire1, errParseInt := strconv.ParseInt(expire, 10, 64)
	if errParseInt != nil {
		fmt.Println("errParseInt:", errParseInt)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	ext := strings.Split(filename, ".")
	prefix := strings.Join(ext[:len(ext)-1], "")

	if ext[len(ext)-1] != "m3u8" {
		temp := strings.Split(prefix, "-")
		prefix = strings.Join(temp[:len(temp)-1], "")
	}
	if videoId != prefix {
		fmt.Println("videoId与密文中的videoId不同")
		sendErrorResponse(rw, models.ErrorVideoIdError)
	}
	if time.Now().Unix() > expire1 {
		fmt.Println("链接已过期")
		sendErrorResponse(rw, models.ErrorExpireError)
		return
	}
	dir, _ := beego.AppConfig.String("FileDir")
	thre, _ := beego.AppConfig.String("threshold")
	threshold, errThreshold := strconv.ParseInt(thre, 10, 64)
	if errThreshold != nil {
		fmt.Println("errThreshold:", errThreshold)
		sendErrorResponse(rw, models.ErrorInternalFaults)
		return
	}
	fmt.Println("threshold:", threshold)
	vl := dir + "/" + prefix + "/" + filename
	fmt.Println("vl:", vl)
	video, err := os.Open(vl)
	if err == errors.ErrNotFound {
		fmt.Println("没有找到视频")
		sendErrorResponse(rw, models.ErrorFileError)
		return
	}
	defer video.Close()
	fileInfo, err := video.Stat()
	if err != nil {
		fmt.Println("Get FileInfo", err.Error())
		sendErrorResponse(rw, models.ErrorFileError)
		return
	}
	// beego.BConfig.WebConfig.ViewsPath
	var size int64

	filesize := fileInfo.Size()
	prexfile := strings.Join(ext[:len(ext)-1], "")
	fmt.Println("prefile:", prexfile)
	key := []byte(address + prexfile)
	var bstr string
	fmt.Println("key:", key)
	mutexDB.Lock()
	value, errGet := db.Get(key, nil)
	if errGet != errors.ErrNotFound && string(value) != "" {
		size, _ = strconv.ParseInt(string(value), 10, 64)
		fmt.Println("size:", size)
		_, bstr := formatFileSize(size)
		log.Printf("该用户的当前视频质押余量为: %s", bstr)
		if ext[len(ext)-1] != "m3u8" {
			if filesize-size > threshold*1024*1024 {
				sendErrorResponse(rw, models.ErrorInsufficientBalance)
				return
			}
		}
	} else {
		size = 0
		fmt.Println("size0:", strconv.FormatInt(size, 10))
		log.Println("该用户的当前视频质押余量为: 0MB")
		if err := db.Put(key, []byte(strconv.FormatInt(size, 10)), nil); err != nil {
			sendErrorResponse(rw, models.ErrorDBError)
		}
	}

	mutexDB.Unlock()
	fmt.Println("接下来;", size)
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
				fmt.Println("sendFile2 start:", start, "end:", end, "size:", fileInfo.Size())
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
		fmt.Println("file locate seek", err.Error())
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	// ext = strings.Split(fileInfo.Name(), ".")
	fmt.Println("ext:", ext[len(ext)-1])

	fmt.Println("content-type:", contentTypeMap[ext[len(ext)-1]])
	if ok := contentTypeMap[ext[len(ext)-1]]; ok != "" {
		rw.Header().Set("Content-Type", contentTypeMap[ext[len(ext)-1]])
	} else {
		rw.Header().Set("Content-Type", "application/octet-stream")
	}
	rw.Header().Add("Accept-Ranges", "bytes")

	// rw.Header().Add("Content-Disposition", "attachment; filename="+fileInfo.Name())
	rw.Header().Add("Content-Disposition", "attachment; filename="+fileInfo.Name())

	defer func(size int64, filesize int64) {
		fmt.Println("fizesize:", filesize)
		fmt.Println("size", size)
		size = size - filesize
		fmt.Println("size:", size)
		mutexDB.Lock()
		fmt.Println("key:", key)
		fmt.Println("sizeformat", strconv.FormatInt(size, 10))
		defer mutexDB.Unlock()
		if errPut := db.Put(key, []byte(strconv.FormatInt(size, 10)), nil); errPut != nil {
			fmt.Println("errPut:", errPut)
			sendErrorResponse(rw, models.ErrorDBError)
		}
		_, sizeUpdate := formatFileSize(size)
		log.Printf("该用户在该视频的余量传输前为%s,传输后余量为：%s", bstr, sizeUpdate)
	}(size, filesize)

	n := 512
	buf := make([]byte, n)
	for {
		if end-start+1 < int64(n) {
			n = int(end - start + 1)
			fmt.Println("n:", n)
		}
		_, err := video.Read(buf[:n])
		if err != nil {
			fmt.Println("1:", err)
			if err != io.EOF {
				fmt.Println("error:", err)
			}
			return
		}
		err = nil
		_, err = rw.Write(buf[:n])
		if err != nil {
			// fmt.Println(err, start, end, info.Size(), n)
			return
		}
		start += int64(n)
		if start >= end+1 {
			return
		}
	}
	fmt.Println("fizesize:", filesize)
	fmt.Println("size", size)
	size = size - filesize

	mutexDB.Lock()
	defer mutexDB.Unlock()
	if errPut := db.Put(key, []byte(strconv.FormatInt(size, 10)), nil); errPut != nil {
		fmt.Println("errPut:", errPut)
		sendErrorResponse(rw, models.ErrorDBError)
	}
	_, sizeUpdate := formatFileSize(size)
	log.Printf("该用户在该视频的余量传输前为%s,传输后余量为：%s", bstr, sizeUpdate)
}

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
	var mutxDB sync.Mutex

	rw := v.Ctx.ResponseWriter

	// 获取签名参数
	paySign := v.GetString("paySign")
	fmt.Println("paySign:", paySign)
	payData := v.GetString("payData")
	fmt.Println("payData:", payData)
	voucherSign, errDecode := parseVoucherSign(paySign)
	if errDecode != nil {
		fmt.Println("errDecode:", errDecode)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	fmt.Println("voucherSign:", voucherSign.VidoId, voucherSign.PayPrivateKey)

	// 对签名参数进行解码
	voucherData, errParsePayData := parsePayData(payData, voucherSign.PayPrivateKey)
	if errParsePayData != nil {
		fmt.Println("errParsePayData", errParsePayData)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	fmt.Println("voucherData:", voucherData.ReceivedSizeMB, voucherData.Timestamp)

	newSize := voucherData.ReceivedSizeMB
	log.Printf("本次支付流量：%sMB", newSize)
	var sizes int64
	mutxDB.Lock()
	key := []byte(voucherSign.Creator + strconv.FormatUint(voucherSign.VidoId, 10))
	data, errGet := db.Get(key, nil)
	mutxDB.Unlock()

	if errGet == errors.ErrNotFound {
		sizes = 0
		log.Printf("用户当前余额是0MB")
	} else {
		if errUnmarshal := json.Unmarshal(data, &sizes); errUnmarshal != nil {
			fmt.Println("errUnmarshal:", errUnmarshal)
			sendErrorResponse(rw, models.ErrorInternalFaults)
		}
		_, str := formatFileSize(sizes)
		log.Printf("用户当前余额是%sMB", str)
	}

	sizes = sizes + int64(newSize)
	fmt.Printf("size:", sizes)
	value, errMarshal := json.Marshal(sizes)
	fmt.Println("value:", value)
	if errMarshal != nil {
		fmt.Println("errMarshal:", errMarshal)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}
	mutxDB.Lock()
	if errPut := db.Put(key, value, nil); errPut != nil {
		fmt.Println("errPut:", errPut)
		sendErrorResponse(rw, models.ErrorDBError)
	}
	_, strPay := formatFileSize(sizes)
	mutxDB.Unlock()
	log.Printf("用户支付后的余额是%sMB", strPay)

	msg, errSubmit := makeSubmitPaysign(voucherSign.Creator, paySign, payData)
	if errSubmit != nil {
		fmt.Println("errSumit:", errSubmit)
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}

	if err := mQueue.Publish("sendTx", msg); err != nil {
		fmt.Println("mq publish error:", err.Error())
		sendErrorResponse(rw, models.ErrorInternalFaults)
	}

	sendNormalResponse(rw, models.SuccessSignRequest, 201)

}

func (v *VideoController) Settlement() {
	var settle *models.SettleInfo
	videoID := v.GetString("videoID")
	fmt.Println("videoID:", videoID)
	if err := v.ParseForm(settle); err != nil {
		fmt.Println("settle parse form:", err)
		return
	}

	v.ServeJSON()
}

func getDecryptMsg(src string) (address string, videoId string, expire string) {
	strs := strings.Split(src, ",")
	address = strings.Split(strs[0], "=")[1]
	videoId = strings.Split(strs[1], "=")[1]
	expire = strings.Split(strs[2], "=")[1]
	return
}

func makeSubmitPaysign(creator string, paySign string, payData string) ([]byte, error) {
	var accountNumber uint64
	var sequence uint64
	priv, err := getPrivKey(creator)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	pub := priv.PubKey()
	addr := sdk.AccAddress(pub.Address())
	accountNumber, sequence, errGetAccount := getAccountNumSequence(addr.String())
	if errGetAccount != nil {
		fmt.Println("errGetAccount:", errGetAccount)
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
	txBuilder.SetGasLimit(200000)
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
		Sequence: sequence,
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

	return []byte(txBytesBase64), nil
}

func getAccountNumSequence(address string) (uint64, uint64, error) {
	remoteAddress, err := beego.AppConfig.String("remoteAddress")
	if err != nil {
		fmt.Println("getremoteAddress:", remoteAddress)
		return 0, 0, err
	}
	req, errRequest := http.NewRequest("GET", fmt.Sprintf("http://%s/cosmos/auth/v1beta1/accounts/%s", remoteAddress, address), strings.NewReader(""))
	if errRequest != nil {
		fmt.Println("errRequest:", errRequest)
		return 0, 0, errRequest
	}
	res, errRes := http.DefaultClient.Do(req)
	if errRes != nil {
		fmt.Println("errRes:", errRes)
		return 0, 0, errRequest
	}
	respBody, err := ioutil.ReadAll(res.Body)
	num, sequence, errParse := parseResNum(string(respBody))
	if errParse != nil {
		fmt.Println("errParse:", errParse)
		return 0, 0, errParse
	}
	fmt.Println("getAccountNumSequence", num, sequence)
	return num, sequence, nil

}

func parseResNum(res string) (uint64, uint64, error) {
	var result models.GetAccountRes
	if err := json.Unmarshal([]byte(res), &result); err != nil {
		return 0, 0, err
	}
	num, _ := strconv.ParseUint(result.Account.AccountNum, 10, 64)
	seq, _ := strconv.ParseUint(result.Account.Sequence, 10, 64)

	return num, seq, nil

}

func getCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func getPrivKey(addr string) (cryptotypes.PrivKey, error) {
	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, err
	}

	kr, err := keyring.New("dataocean", keyring.BackendTest, "~/.dataocean", nil, getCodec())
	k, err := kr.KeyByAddress(acc)
	if err != nil {
		return nil, err
	}

	rl := k.GetLocal()
	privKey := rl.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
	return privKey, nil
}

func parsePayData(payData string, privateKey string) (*models.VoucherPayData, error) {
	voucherData := &models.VoucherPayData{}
	decrptData := dongle.Decrypt.FromBase64String(payData).ByRsa(privateKey).ToString()
	if err := json.Unmarshal([]byte(decrptData), voucherData); err != nil {
		fmt.Println("err:", err)
		return nil, err
	}

	return voucherData, nil
}

func parseVoucherSign(paySign string) (*models.VoucherPaySign, error) {
	voucherSign := &models.VoucherPaySign{}
	parseSign, err := parsePaySign(sdk.Context{}, paySign)
	if err != nil {
		fmt.Println("parsePaySign:", err)
		return nil, err
	}
	voucherSign.PayPrivateKey = parseSign.PayPrivateKey
	voucherSign.VidoId = parseSign.VideoId
	voucherSign.Creator = parseSign.Creator

	return voucherSign, nil

}

func parsePaySign(ctx sdk.Context, paySignStr string) (*cosTypes.MsgPaySign, error) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()

	std.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &cosTypes.MsgPaySign{})
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txConfig := tx.NewTxConfig(protoCodec, tx.DefaultSignModes)

	// txBytes, err := base64.StdEncoding.DecodeString(paySignStr)
	txBytes, err := hex.DecodeString(paySignStr)

	// txBytes, err := hex.DecodeString(paySignStr)
	if err != nil {
		return nil, err
	}
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

func formatFileSize(fileSize int64) (size float64, sizstr string) {
	var str string
	if fileSize < 1024 || fileSize > -1024 {
		str = fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
		return float64(fileSize) / float64(1), str
	} else if fileSize < (1024*1024) || fileSize > -(1024*1024) {
		str = fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
		return float64(fileSize) / float64(1024), str
	} else if fileSize < (1024*1024*1024) || fileSize > -(1024*1024*1024) {
		str = fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
		return float64(fileSize) / float64(1024*1024), str
	} else if fileSize < (1024*1024*1024*1024) || fileSize > -(1024*1024*1024*1024) {
		str = fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
		return float64(fileSize) / float64(1024*1024*1024), str
	} else if fileSize < (1024*1024*1024*1024*1024) || fileSize > -(1024*1024*1024*1024*1024) {
		str = fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
		return float64(fileSize) / float64(1024*1024*1024*1024), str
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		str = fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
		return float64(fileSize) / float64(1024*1024*1024*1024*1024), str
	}
}
