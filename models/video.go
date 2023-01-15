package models

import (
	"errors"
)

type Video struct {
	Id string `json:"id"`
	// AuthorId int	`json:"author_id"`
	Name string `json:"name"`

	PicUrl string `json:"pic_url"`

	Price string `json:"price"`

	Duration int `json:"duration"`

	Size int `json:"size"`

	// DisplayCtime string		`json:"display_ctime"`
}

type VoucherInfo struct {
	Type           string `json:"@type"`
	Creator        string `json:"creator"`
	VidoId         uint64 `json:"videoId"`
	Level          uint64 `json:"sn"`
	Sn             uint64 `json:"size"`
	ReceivedSizeMB uint64 `json:"receivedSizeMB"`
	Timestamp      uint64 `json:"paytime"`
}

// type VoucherSign struct {
//	Creator string `json:"creator" protobuf:"bytes,1,opt,name=creator,json=creator,proto3"`
//	Sign    string `json:"paySign" protobuf:"bytes,2,opt,name=paySign,json=paySign,proto3"`
// }

type SignBody struct {
	Messages                    Messages      `json:"message"`
	Memo                        interface{}   `json:"memo"`
	Timeoutheight               uint64        `json:"timeout_height"`
	Extensionoptions            []interface{} `json:"extension_options"`
	NonCriticalExtensionOptions []interface{} `json:"non_critical_extension_options"`
}
type GetAccountRes struct {
	Account struct {
		AccountNum string `json:"account_number"`
		Sequence   string `json:"sequence"`
	} `json:"account"`
}

type Account struct {
	Type    string `json:"@type"`
	Address string `json:"address"`
	Pubkey  PubKey `json:"pub_key"`
}

type PubKey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type Publickey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type VoucherPayData struct {
	ReceivedSizeMB float64 `json:"receivedSizeMB"`
	Timestamp      uint64  `json:"timestamp"`
}

type VoucherPaySign struct {
	Creator       string `json:"creator"`
	VidoId        uint64 `json:"videoId"`
	PayPrivateKey string `json:"payPrivateKey"`
}

type ModeInfo struct {
	Single string `json:"single"`
}

type Messages struct {
	MessageNums []VoucherInfo
}

type Err struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}

type ErrResponse struct {
	HttpSC int
	Error  Err
}

type NormalResponse struct {
	HttpSc  int
	Message string
}

type SettleInfo struct {
	UserAddress string `json:"user_address"`
	Charge      string `json:"charge"`
	TimeStamp   string `json:"time_stamp"`
}
type SettleResponse struct {
	Result string `json:"result"`
	msg    string `json:"msg"`
	error  string `json:"error"`
}

var (
	VideoList map[string]*Video
)

var (
	SuccessSignRequest          = NormalResponse{HttpSc: 201, Message: "接收签名成功"}
	ErrorRequestBodyParseFailed = ErrResponse{HttpSC: 400, Error: Err{Error: "Request body is not correct", ErrorCode: "001"}}
	ErrorInsufficientBalance    = ErrResponse{HttpSC: 401, Error: Err{Error: "User's balance is not enough.", ErrorCode: "002"}}
	ErrorDBError                = ErrResponse{HttpSC: 500, Error: Err{Error: "DB ops failed", ErrorCode: "003"}}
	ErrorInternalFaults         = ErrResponse{HttpSC: 500, Error: Err{Error: "Internal service error", ErrorCode: "004"}}
	ErrorChainError             = ErrResponse{HttpSC: 500, Error: Err{Error: "Chain interaction error", ErrorCode: "005"}}
	ErrorFileError              = ErrResponse{HttpSC: 500, Error: Err{Error: "operate file error", ErrorCode: "006"}}
	ErrorBadRequestError        = ErrResponse{HttpSC: 400, Error: Err{Error: "Bad request error", ErrorCode: "007"}}
	ErrorVideoIdError           = ErrResponse{HttpSC: 400, Error: Err{Error: "请求参数中的videoID与密文中的不一致", ErrorCode: "008"}}
	ErrorExpireError            = ErrResponse{HttpSC: 400, Error: Err{Error: "请求链接已过期", ErrorCode: "009"}}
	ErrorValidateError          = ErrResponse{HttpSC: 400, Error: Err{Error: "签名验证失败", ErrorCode: "010"}}
)

func AddVideo(video *Video) string {
	VideoList[video.Name] = video
	return video.Id
}

func GetVideo(vname string) (*Video, error) {
	if v, ok := VideoList[vname]; ok {
		return v, nil
	}
	return nil, errors.New("Video not exists")
}

func GetAllVideo() map[string]*Video {
	return VideoList
}

func UpdateVideo(vname string, video *Video) (*Video, error) {
	if vv, ok := VideoList[vname]; ok {
		if video.Id != "" {
			vv.Id = video.Id
		}
		if video.Name != "" {
			vv.Name = video.Name
		}
		if video.PicUrl != "" {
			vv.PicUrl = video.Name
		}
		if video.Size != 0 {
			vv.Size = video.Size
		}
		if video.Duration != 0 {
			vv.Duration = video.Duration
		}
		if video.Price != "" {
			vv.Price = video.Price
		}
		return vv, nil

	}
	return nil, errors.New("video not exists")
}
