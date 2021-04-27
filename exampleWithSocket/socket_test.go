package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"log"
	"testing"

	bililive "github.com/BaoJW/douyu-barrage"
)

func TestSocket(t *testing.T) {
	ip := "openapi-danmu.douyu.com"
	port := 80
	flag.Parse()

	socket := newSocket(ip, port)

	live := &bililive.Live{
		ReceiveGift: func(roomID int, gift *bililive.GiftModel) {
			log.Printf("【礼物】%v:  %v(%v) * %v  价格 %v  连击 %v", gift.UserName, gift.GiftName, gift.GiftID, gift.Num, gift.Price, gift.Combo)
			m := msg{
				ID:    gift.GiftID,
				Name:  gift.GiftName,
				Count: gift.Num,
				User:  gift.UserName,
				Price: gift.Price,
			}
			body, _ := json.Marshal(&m)
			buff := bytes.NewBuffer([]byte{})
			_ = binary.Write(buff, binary.BigEndian, int16(len(body)))
			_ = binary.Write(buff, binary.LittleEndian, body)
			socket.sendTCP(buff.Bytes())
		},
		ReceivePopularValue: func(roomID int, value uint32) {
			log.Printf("【人气】:  %v", value)
		},
	}
	live.Start(context.TODO())
	_ = live.Join(288016)
	scanner(socket)
}
