# douyu-barrage
斗鱼弹幕服务器接入--golang实现

### 功能
1. 接入斗鱼弹幕服务器，获取直播间弹幕、礼物、动态消息等

### 安装
```asciidoc
go get -u github.com/BaoJW/douyu-barrage
```

### aid和secret获取
```asciidoc
需要去斗鱼申请接入的开发者权限
注意：若申请权限就没有办法获取token,从而无法接入斗鱼弹幕服务器
https://open.douyu.com/manage/ 
```

### 斗鱼文档
```asciidoc
https://open.douyu.com/source/
```

### 快速开始
```asciidoc
func main() {
	aid := flag.String("aid", "", "aid")
	secret := flag.String("secret", "", "secret")
	roomID := flag.Int("roomId", 0, "roomId")
	ip := flag.String("ip", "", "ip")
	port := flag.Int("port", 0, "port")

	flag.Parse()

	if *aid == "" {
		log.Fatalln("aid不能为空!")
		return
	}

	if *secret == "" {
		log.Fatalln("secret不能为空!")
		return
	}

	if *roomID <= 0 {
		log.Fatalln("房间号错误!")
		return
	}

	//远程获取pprof数据
	go func() {
		log.Println(http.ListenAndServe("localhost:8080", nil))
	}()

	live := &douyulive.Live{
		Debug:              false, // 不输出日志
		AnalysisRoutineNum: 1,     // 消息分析协程数量，默认为1，为1可以保证通知顺序与接收到消息顺序相同
		LoginRespMessageHandler: func(roomID int, msg *douyulive.LoginRespMessageModel) {
			log.Printf("【登录消息】%s 登录成功", msg.NickName)
		},
		BarrageMessageHandler: func(roomID int, msg *douyulive.BarrageMessageModel) {
			log.Printf("【弹幕消息】%s 说：%s", msg.NickName, msg.Txt)
		},
		StormMessageHandler: func(roomID int, msg *douyulive.StormMessage) {
			log.Printf("【领取在线鱼丸消息】%s 在 %d 直播间领取%d鱼丸", msg.NickName, msg.RoomID, msg.Sil)
		},
		SendGiftMessageHandler: func(roomID int, msg *douyulive.SendGiftMessage) {
			log.Printf("【赠送礼物消息】%s在%d直播间赠送%d礼物，礼物ID: %d", msg.NickName, msg.RoomID, msg.GfCount, msg.GiftID)
		},
		SpecialUserMessageHandler: func(roomID int, msg *douyulive.SpecialUserMessage) {
			log.Printf("【用户进房通知消息】欢迎 %s 进入直播间%d", msg.NickName, msg.RoomID)
		},
		SwitchBroadcastMessageHandler: func(roomID int, msg *douyulive.SwitchBroadcastMessage) {
			if msg.Status == 0 {
				log.Printf("【直播间开关播提醒消息】 %d直播间当前没有直播", msg.RoomID)
			} else if msg.Status == 1 {
				log.Printf("【直播间开关播提醒消息】 %d直播间当前正在直播", msg.RoomID)
			}

		},
		BroadcastRankMessageHandler: func(roomID int, msg *douyulive.BroadcastRankMessage) {
			log.Printf("【排行榜消息】 %d直播间当前时间%s排行榜消息 总榜: %v, 周榜: %v, 日榜: %v", msg.RoomID, time.Unix(msg.Timestamp, 0), msg.ListAll, msg.List, msg.ListDay)
		},
		SuperBarrageMessageHandler: func(roomID int, msg *douyulive.SuperBarrageMessage) {
			log.Printf("【超级弹幕消息】 %d直播间超级弹幕消息: %s", msg.RoomID, msg.Content)
		},
		RoomGiftBarrageMessageHandler: func(roomID int, msg *douyulive.RoomGiftBroadcastMessage) {
			log.Printf("【房间内礼物广播消息】 %d直播间 %s赠送给 %s %d个 %d", msg.RoomID, msg.SendNickName, msg.DoneeNickName, msg.GiftCount, msg.GiftID)
		},
	}
	live.Start(context.Background())
	_ = live.Join(*aid, *secret, *ip, *port, *roomID)
	live.Wait()
}

```

### 最后
```asciidoc
各位老爷们如果觉得好用，就给小的一个star吧
```
