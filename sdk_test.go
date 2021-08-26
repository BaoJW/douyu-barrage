package douyulive

import (
	"context"
	"log"
	"sort"
	"testing"
)

const aid = "xxx"
const secret = "xxx"

func TestLive_Start(t *testing.T) {
	live := &Live{
		Debug:              false, // 不输出日志
		AnalysisRoutineNum: 1,     // 消息分析协程数量，默认为1，为1可以保证通知顺序与接收到消息顺序相同
		LoginRespMessageHandler: func(roomID int, msg *LoginRespMessageModel) {
			log.Printf("【登录消息】%s 登录成功", msg.NickName)
		},
		BarrageMessageHandler: func(roomID int, msg *BarrageMessageModel) {
			log.Printf("【弹幕消息】%s 说：%s", msg.NickName, msg.Txt)
		},
		StormMessageHandler: func(roomID int, msg *StormMessage) {
			log.Printf("【领取在线鱼丸消息】%s 在 %d 直播间领取%d鱼丸", msg.NickName, msg.RoomID, msg.Sil)
		},
		SendGiftMessageHandler: func(roomID int, msg *SendGiftMessage) {
			log.Printf("【赠送礼物消息】%s在%d直播间赠送%d礼物，礼物ID: %d", msg.NickName, msg.RoomID, msg.GfCount, msg.GiftID)
		},
		SpecialUserMessageHandler: func(roomID int, msg *SpecialUserMessage) {
			log.Printf("【用户进房通知消息】欢迎 %s 进入直播间%d", msg.NickName, msg.RoomID)
		},
		SwitchBroadcastMessageHandler: func(roomID int, msg *SwitchBroadcastMessage) {
			if msg.Status == 0 {
				log.Printf("【直播间开关播提醒消息】 %d直播间当前没有直播", msg.RoomID)
			} else if msg.Status == 1 {
				log.Printf("【直播间开关播提醒消息】 %d直播间当前正在直播", msg.RoomID)
			}

		},
		BroadcastRankMessageHandler: func(roomID int, msg *BroadcastRankMessage) {
			log.Printf("【排行榜消息】 %d直播间当前排行榜消息: ", msg.RoomID)
			sort.SliceStable(msg.ListAll, func(i, j int) bool {
				return msg.ListAll[i].CurrentRank < msg.ListAll[j].CurrentRank
			})

			log.Print("总榜: ")
			for _, list := range msg.ListAll {
				log.Printf("用户: %s 排行榜第%d", list.NickName, list.CurrentRank)
			}

			log.Print("周榜: ")
			for _, list := range msg.List {
				log.Printf("用户: %s 排行榜第%d", list.NickName, list.CurrentRank)
			}

			log.Print("日榜: ")
			for _, list := range msg.ListDay {
				log.Printf("用户: %s 排行榜第%d", list.NickName, list.CurrentRank)
			}

		},
		SuperBarrageMessageHandler: func(roomID int, msg *SuperBarrageMessage) {
			log.Printf("【超级弹幕消息】 %d直播间超级弹幕消息: %s", msg.RoomID, msg.Content)
		},
		RoomGiftBarrageMessageHandler: func(roomID int, msg *RoomGiftBroadcastMessage) {
			log.Printf("【房间内礼物广播消息】 %d直播间 %s赠送给 %s %d个 %d", msg.RoomID, msg.SendNickName, msg.DoneeNickName, msg.GiftCount, msg.GiftID)
		},
	}
	ctx := context.Background()
	live.Start(ctx)
	_ = live.Join(aid, secret, "", 0, 288016)
	go live.ReJoin(ctx)
	live.Wait()
}
