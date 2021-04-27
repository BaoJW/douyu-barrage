package douyulive

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	LoginRespType   = "loginresp"
	BarrageRespType = "chatmsg"
)

// Start 开始接收
func (live *Live) Start(ctx context.Context) {
	live.ctx = ctx

	rand.Seed(time.Now().Unix())
	if live.AnalysisRoutineNum <= 0 {
		live.AnalysisRoutineNum = 1
	}

	live.room = make(map[int]*liveRoom)
	live.chSocketMessage = make(chan *socketMessage, 30)
	live.chOperation = make(chan *operateInfo, 300)

	live.wg = sync.WaitGroup{}

	for i := 0; i < live.AnalysisRoutineNum; i++ {
		live.wg.Add(1)
		go func() {
			defer live.wg.Done()
			live.analysis(ctx)
		}()
	}

	live.wg.Add(1)
	go func() {
		defer live.wg.Done()
		live.split(ctx)
	}()
}

func (live *Live) Wait() {
	live.wg.Wait()
}

// Join 添加房间
func (live *Live) Join(roomIDs ...int) error {
	if len(roomIDs) == 0 {
		return errors.New("没有要添加的房间")
	}

	for _, roomID := range roomIDs {
		if _, exist := live.room[roomID]; exist {
			return fmt.Errorf("房间 %d 已存在", roomID)
		}
	}
	for _, roomID := range roomIDs {
		nextCtx, cancel := context.WithCancel(live.ctx)

		room := &liveRoom{
			roomID: roomID,
			cancel: cancel,
		}
		live.room[roomID] = room
		room.enter()
		go room.heartBeat(nextCtx)
		go room.receive(nextCtx, live.chSocketMessage)
	}
	return nil
}

// Remove 移出房间
func (live *Live) Remove(roomIDs ...int) error {
	if len(roomIDs) == 0 {
		return errors.New("没有要移出的房间")
	}

	for _, roomID := range roomIDs {
		if room, exist := live.room[roomID]; exist {
			room.cancel()
			delete(live.room, roomID)
		}
	}
	return nil
}

// 拆分数据
func (live *Live) split(ctx context.Context) {
	var (
		message *socketMessage
	)
	for {
		message = <-live.chSocketMessage
		for len(message.body) > 0 {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg := ByteToMsg(message.body)
			if _, ok := msg["type"]; !ok {
				log.Println("数据格式不匹配", msg)
				continue
			}

			live.chOperation <- &operateInfo{RoomID: message.roomID, Type: msg["type"], Buffer: msg}
		}
	}
}

// 分析接收到的数据
func (live *Live) analysis(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		buffer := <-live.chOperation
		switch buffer.Type {
		case LoginRespType:
			live.room[buffer.RoomID].joinGroup(time.Now())
		case BarrageRespType:
			if live.ReceiveMsg != nil {
				live.ReceiveMsg(buffer.RoomID, buffer.Buffer)
			}
		}
	}

	//case WS_OP_CONNECT_SUCCESS:
	//	if live.Debug {
	//		log.Println("CONNECT_SUCCESS", string(buffer.Buffer))
	//	}
	//case WS_OP_MESSAGE:
	//	result := cmdModel{}
	//	err := json.Unmarshal(buffer.Buffer, &result)
	//	if err != nil {
	//		if live.Debug {
	//			log.Println(err)
	//			log.Println(string(buffer.Buffer))
	//		}
	//		continue
	//	}
	//	temp, err := json.Marshal(result.Data)
	//	if err != nil {
	//		if live.Debug {
	//			log.Println(err)
	//		}
	//		continue
	//	}
	//	switch result.CMD {
	//	case "LIVE": // 直播开始
	//		log.Println(string(buffer.Buffer))
	//		if live.Live != nil {
	//			live.Live(buffer.RoomID)
	//		}
	//	case "CLOSE": // 关闭
	//		fallthrough
	//	case "PREPARING": // 准备
	//		fallthrough
	//	case "END": // 结束
	//		log.Println(string(buffer.Buffer))
	//		if live.End != nil {
	//			live.End(buffer.RoomID)
	//		}
	//	case "SYS_MSG": // 系统消息
	//		if live.SysMessage != nil {
	//			m := &SysMsgModel{}
	//			_ = json.Unmarshal(buffer.Buffer, m)
	//			live.SysMessage(buffer.RoomID, m)
	//		}
	//	case "ROOM_CHANGE": // 房间信息变更
	//		if live.RoomChange != nil {
	//			m := &RoomChangeModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.RoomChange(buffer.RoomID, m)
	//		}
	//	case "WELCOME": // 用户进入
	//		if live.UserEnter != nil {
	//			m := &UserEnterModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.UserEnter(buffer.RoomID, m)
	//		}
	//	case "WELCOME_GUARD": // 舰长进入
	//		if live.GuardEnter != nil {
	//			m := &GuardEnterModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.GuardEnter(buffer.RoomID, m)
	//		}
	//	case "DANMU_MSG": // 弹幕
	//		if live.ReceiveMsg != nil {
	//			msgContent := result.Info[1].(string)
	//
	//			if live.StormFilter && live.storming[buffer.RoomID] {
	//				for _, value := range live.stormContent[buffer.RoomID] {
	//					if msgContent == value {
	//						//log.Println("过滤弹幕：", value)
	//						continue analysis
	//					}
	//				}
	//			}
	//
	//			userInfo := result.Info[2].([]interface{})
	//			medalInfo := result.Info[3].([]interface{})
	//			m := &MsgModel{
	//				UserID:    int64(userInfo[0].(float64)),
	//				UserName:  userInfo[1].(string),
	//				UserLevel: int(result.Info[4].([]interface{})[0].(float64)),
	//				Content:   msgContent,
	//				Timestamp: int64(result.Info[9].(map[string]interface{})["ts"].(float64)),
	//			}
	//			if len(medalInfo) >= 4 {
	//				m.MedalLevel = int(medalInfo[0].(float64))
	//				m.MedalName = medalInfo[1].(string)
	//				m.MedalUpName = medalInfo[2].(string)
	//				m.MedalRoomID = int64(medalInfo[3].(float64))
	//			}
	//			live.ReceiveMsg(buffer.RoomID, m)
	//		}
	//	case "SEND_GIFT": // 礼物通知
	//		if live.ReceiveGift != nil {
	//			m := &GiftModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.ReceiveGift(buffer.RoomID, m)
	//		}
	//	case "COMBO_SEND": // 连击
	//		if live.GiftComboSend != nil {
	//			m := &ComboSendModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.GiftComboSend(buffer.RoomID, m)
	//		}
	//	case "COMBO_END": // 连击结束
	//		if live.GiftComboEnd != nil {
	//			m := &ComboEndModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.GiftComboEnd(buffer.RoomID, m)
	//		}
	//	case "GUARD_BUY": // 上船
	//		if live.GuardBuy != nil {
	//			m := &GuardBuyModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.GuardBuy(buffer.RoomID, m)
	//		}
	//	case "ROOM_REAL_TIME_MESSAGE_UPDATE": // 粉丝数更新
	//		if live.FansUpdate != nil {
	//			m := &FansUpdateModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.FansUpdate(buffer.RoomID, m)
	//		}
	//	case "ROOM_RANK": // 小时榜
	//		if live.RoomRank != nil {
	//			m := &RankModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.RoomRank(buffer.RoomID, m)
	//		}
	//	case "SPECIAL_GIFT": // 特殊礼物
	//		m := &SpecialGiftModel{}
	//		_ = json.Unmarshal(temp, m)
	//		if m.Storm.Action == "start" {
	//			m.Storm.ID, _ = strconv.ParseInt(m.Storm.TempID.(string), 10, 64)
	//		}
	//		if m.Storm.Action == "end" {
	//			m.Storm.ID = int64(m.Storm.TempID.(float64))
	//		}
	//		if live.StormFilter && live.ReceiveMsg != nil {
	//			if m.Storm.Action == "start" {
	//				live.storming[buffer.RoomID] = true
	//				live.stormContent[buffer.RoomID][m.Storm.ID] = m.Storm.Content
	//				//log.Println("添加过滤弹幕：", m.Storm.ID, m.Storm.Content)
	//			}
	//			if m.Storm.Action == "end" {
	//				delete(live.stormContent[buffer.RoomID], m.Storm.ID)
	//				live.storming[buffer.RoomID] = len(live.stormContent) > 0
	//				//log.Println("移除过滤弹幕：", m.Storm.ID, live.storming)
	//			}
	//		}
	//		if live.SpecialGift != nil {
	//			live.SpecialGift(buffer.RoomID, m)
	//		}
	//	case "SUPER_CHAT_MESSAGE": // 醒目留言
	//		if live.SuperChatMessage != nil {
	//			m := &SuperChatMessageModel{}
	//			_ = json.Unmarshal(temp, m)
	//			live.SuperChatMessage(buffer.RoomID, m)
	//		}
	//	case "SUPER_CHAT_MESSAGE_JPN":
	//		if live.Debug {
	//			log.Println(string(buffer.Buffer))
	//		}
	//	case "SYS_GIFT": // 系统礼物
	//		fallthrough
	//	case "BLOCK": // 未知
	//		fallthrough
	//	case "ROUND": // 未知
	//		fallthrough
	//	case "REFRESH": // 刷新
	//		fallthrough
	//	case "ACTIVITY_BANNER_UPDATE_V2": //
	//		fallthrough
	//	case "ANCHOR_LOT_CHECKSTATUS": //
	//		fallthrough
	//	case "GUARD_MSG": // 舰长信息
	//		fallthrough
	//	case "NOTICE_MSG": // 通知信息
	//		fallthrough
	//	case "GUARD_LOTTERY_START": // 舰长抽奖开始
	//		fallthrough
	//	case "USER_TOAST_MSG": // 用户通知消息
	//		fallthrough
	//	case "ENTRY_EFFECT": // 进入效果
	//		fallthrough
	//	case "WISH_BOTTLE": // 许愿瓶
	//		fallthrough
	//	case "ROOM_BLOCK_MSG":
	//		fallthrough
	//	case "WEEK_STAR_CLOCK":
	//		fallthrough
}

func (room *liveRoom) createConnect() {
	for {
		if room.server == "" || room.port == 0 {
			room.server = "openapi-danmu.douyu.com"
			room.port = 80
		}

		counter := 0
		for {
			log.Println("尝试创建连接：", room.server, room.port)
			conn, err := connect(room.server, room.port)
			if err != nil {
				log.Println("connect err:", err)
				if counter == 3 {
					log.Printf("尝试创建连接失败: %s", err.Error())
					break
				}
				time.Sleep(1 * time.Second)
				counter++
				continue
			}
			room.conn = conn
			log.Println("连接创建成功：", room.server, room.port)
			return
		}
	}
}

func (room *liveRoom) enter() {
	room.createConnect()

	currentTime := time.Now()
	if room.token == "" || currentTime.Unix()-room.tokenTime >= 60*60*2 {
		token, err := HttpGetDouYuToken(room.aid, room.secret, currentTime)
		if err != nil {
			log.Panic(err)
		}
		room.token = string(token)
		room.tokenTime = currentTime.Unix()
	}

	room.login(currentTime)

}

// 登录
func (room *liveRoom) login(currentTime time.Time) {
	auth := Md5(fmt.Sprintf("%s_%s_%d_%s", room.secret, room.aid, currentTime.Unix(), room.token))

	loginMessage := MsgToByte(map[string]string{"type": "loginreq", "roomid": strconv.Itoa(room.realRoomID), "aid": room.aid, "token": room.token, "time": strconv.Itoa(int(currentTime.Unix())), "auth": auth})

	// 登录弹幕服务器
	if _, err := room.conn.Write(loginMessage); err != nil {
		log.Panic("login failed:", err)
	}

	room.auth = auth
}

// 入组
func (room *liveRoom) joinGroup(currentTime time.Time) {
	// 加入组
	if _, err := room.conn.Write(MsgToByte(map[string]string{
		"type":  "joingroup",
		"rid":   "288016",
		"token": room.token,
		"time":  strconv.Itoa(int(currentTime.Unix())),
		"auth":  room.auth,
	})); err != nil {
		log.Panic("joinGroup failed:", err)
	}

}

// 心跳
func (room *liveRoom) heartBeat(ctx context.Context) {
	var errorCount = 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if _, err := room.conn.Write(MsgToByte(map[string]string{
			"type": "mrkl",
		})); err != nil {
			if errorCount > 10 {
				break
			}
			log.Printf("heatbeat failed")
			errorCount++
		}
		errorCount = 0
		time.Sleep(45 * time.Second)
	}
}

// 接收消息
func (room *liveRoom) receive(ctx context.Context, chSocketMessage chan<- *socketMessage) {
	// 包头总长12个字节
	headerBuffer := make([]byte, HeadLen*2+MsgTypeLen+KeepLen)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// 读取协议头
		_, err := room.conn.Read(headerBuffer)
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Println("read err:", err)
			continue
		}

		// 包体
		var messageBody = make([]byte, int(binary.LittleEndian.Uint32(headerBuffer[0:4]))-int(HeadLen+MsgTypeLen+KeepLen))
		_, err = room.conn.Read(messageBody)
		if err != nil {
			log.Println("read err:", err)
			continue
		}

		chSocketMessage <- &socketMessage{
			roomID: room.roomID,
			body:   messageBody,
		}
	}
}

// 发送数据
func (room *liveRoom) sendData(operation int32, payload []byte) {

	b := bytes.NewBuffer([]byte{})
	head := messageHeader{
		Length:          int32(len(payload)) + WS_PACKAGE_HEADER_TOTAL_LENGTH,
		HeaderLength:    int16(WS_PACKAGE_HEADER_TOTAL_LENGTH),
		ProtocolVersion: WS_HEADER_DEFAULT_VERSION,
		Operation:       operation,
		SequenceID:      WS_HEADER_DEFAULT_SEQUENCE,
	}
	err := binary.Write(b, binary.BigEndian, head)
	if err != nil {
		log.Println(err)
	}

	err = binary.Write(b, binary.LittleEndian, payload)
	if err != nil {
		log.Println(err)
	}

	_, err = room.conn.Write(b.Bytes())
	if err != nil {
		log.Println(err)
	}
}

func connect(host string, port int) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	return net.DialTCP("tcp", nil, tcpAddr)
}

// 进行zlib解压缩
func doZlibUnCompress(compressSrc []byte) []byte {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		log.Println("zlib", err)
	}
	_, err = io.Copy(&out, r)
	if err != nil {
		log.Println("zlib copy", err)
	}
	return out.Bytes()
}
