# douyu-barrage
斗鱼弹幕TCP获取

### 功能
1. 获取斗鱼服务器弹幕消息，礼物消息，动态消息等

### 安装
```asciidoc
go get -u github.com/BaoJW/douyu-barrage
```

### aid和secret获取
```asciidoc
https://open.douyu.com/manage/ 
要去斗鱼的申请接入的开发者权限,请注意没有申请权限就没有办法获取token,从而无法接入斗鱼弹幕服务器
```

### 斗鱼文档
```asciidoc
https://open.douyu.com/source/
```

### 快速开始
```asciidoc
	aid := flag.String("aid", "", "aid")
	secret := flag.String("secret", "", "secret")
	roomID := flag.Int("id", 0, "id")
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
	}
	live.Start(context.Background())
	_ = live.Join(*aid, *secret, *roomID)
	live.Wait()
```

### 最后
```asciidoc
各位老爷们如果觉得好用，就给小的一个star吧
```
