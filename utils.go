package douyulive

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// md5哈希加密
func Md5(key string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(key))
	return hex.EncodeToString(md5Ctx.Sum(nil))
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

func HttpGetDouYuToken(aid, secret string, currentTime time.Time) ([]byte, error) {
	auth := Md5(fmt.Sprintf("/api/thirdPart/token?aid=%s&time=%d%s", aid, currentTime.Unix(), secret))

	url := fmt.Sprintf("https://openapi.douyu.com/api/thirdPart/token?aid=%s&time=%d&auth=%s", aid, currentTime.Unix(), auth)
	resp, err := httpSend(url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
