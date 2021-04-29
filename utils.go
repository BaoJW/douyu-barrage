package douyulive

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// md5哈希加密
func Md5(key string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(key))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

func StrToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func httpSend(url string) ([]byte, error) {
	tr := &http.Transport{ //解决x509: certificate signed by unknown authority
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func httpGetDouYuToken(aid, secret string, currentTime time.Time) ([]byte, error) {
	auth := Md5(fmt.Sprintf("/api/thirdPart/token?aid=%s&time=%d%s", aid, currentTime.Unix(), secret))

	url := fmt.Sprintf("https://openapi.douyu.com/api/thirdPart/token?aid=%s&time=%d&auth=%s", aid, currentTime.Unix(), auth)
	resp, err := httpSend(url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GenerateToken(aid, secret string, currentTime time.Time) (string, error) {
	resp, err := httpGetDouYuToken(aid, secret, currentTime)
	if err != nil {
		return "", err
	}

	douYuTokenResp, err := MarshalDouYuData(resp, &TokenInfo{})
	if err != nil {
		return "", err
	}

	return douYuTokenResp.(*TokenInfo).Token, nil

}

// 斗鱼接口数据返回
type DouYuResponse struct {
	Code int64       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Token返回信息
type TokenInfo struct {
	Token  string `json:"token"`
	Expire int    `json:"expire"` // 过期时间，2小时
}

// 弹幕信息
type BarrageInfo struct {
	List        []*BarrageList `json:"list"`
	Count       int64          `json:"cnt"`
	PageContext int64          `json:"page_context"`
}

type BarrageList struct {
	RoomID    int64  `json:"room_id"`   // 房间号
	UID       int64  `json:"uid"`       // 用户uid
	Nickname  string `json:"nickname"`  // 主播昵称
	Content   string `json:"content"`   // 弹幕内容
	Timestamp int64  `json:"timestamp"` // 发送时间戳
	Ip        string `json:"ip"`        // 用户IP
	Platform  string `json:"platform"`  // 直播平台
}

// 直播间信息
type RoomInfo struct {
	RoomID   int64  `json:"rid"`       // 房间号
	Attendee int64  `json:"hn"`        // 人气值
	RoomName string `json:"room_name"` // 房间名
}

func MarshalDouYuData(resp []byte, module interface{}) (interface{}, error) {
	douYuResp := new(DouYuResponse)
	if err := json.Unmarshal(resp, douYuResp); err != nil {
		return nil, err
	}

	d, err := json.Marshal(douYuResp.Data)
	if err != nil {
		return nil, err
	}

	switch module.(type) {
	case *RoomInfo:
		douYuRoomInfo := new(RoomInfo)
		if err := json.Unmarshal(d, douYuRoomInfo); err != nil {
			return nil, err
		}
		return douYuRoomInfo, nil
	case *TokenInfo:
		tokenInfo := new(TokenInfo)
		if err := json.Unmarshal(d, tokenInfo); err != nil {
			return nil, err
		}
		return tokenInfo, nil
	case *BarrageInfo:
		douYuBarrageInfo := new(BarrageInfo)
		if err := json.Unmarshal(d, douYuBarrageInfo); err != nil {
			return nil, err
		}

		return douYuBarrageInfo, nil

	default:
		return nil, errors.New("unsupported module type")
	}

}
