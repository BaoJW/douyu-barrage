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
	LoginRespType             = "loginresp"
	BarrageRespType           = "chatmsg"
	StormRespType             = "onlinegift"
	SendGiftRespType          = "dgb"
	SpecialUserRespType       = "uenter"
	SwitchBroadcastRespType   = "rss"
	BroadcastRankRespType     = "ranklist"
	SuperBarrageRespType      = "ssd"
	RoomGiftBroadcastRespType = "spbc"
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
	chReconSignal = make(chan *liveRoom, 1)

	live.wg = sync.WaitGroup{}

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
func (live *Live) Join(aid, secret, ip string, port int, roomIDs ...int) error {
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
			aid:    aid,
			secret: secret,
			server: ip,
			port:   port,
		}
		live.room[roomID] = room
		room.enter()
		go room.heartBeat(nextCtx)
		go room.receive(nextCtx, live.chSocketMessage)
	}
	return nil
}

// 房间异常断开后的重连
func (live *Live) ReJoin(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("尝试重新连接失败：", err)
			return
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case liveRoomInfo := <-chReconSignal:
			if liveRoomInfo.reconnect {
				nextCtx, cancel := context.WithCancel(live.ctx)

				room := &liveRoom{
					roomID: liveRoomInfo.roomID,
					cancel: cancel,
					aid:    liveRoomInfo.aid,
					secret: liveRoomInfo.secret,
					server: liveRoomInfo.server,
					port:   liveRoomInfo.port,
				}
				live.room[liveRoomInfo.roomID] = room
				room.enter()
				go room.heartBeat(nextCtx)
				go room.receive(nextCtx, live.chSocketMessage)
			}

		default:
		}

	}
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

			switch message.body["type"] {
			case LoginRespType:
				live.room[message.roomID].joinGroup()
				if live.LoginRespMessageHandler != nil {
					live.LoginRespMessageHandler(message.roomID, TransferLoginRespMessage(message.body))
				}
			case BarrageRespType:
				if live.BarrageMessageHandler != nil {
					live.BarrageMessageHandler(message.roomID, TransferBarrageMessage(message.body))
				}
			case StormRespType:
				if live.StormMessageHandler != nil {
					live.StormMessageHandler(message.roomID, TransferStormMessage(message.body))
				}
			case SendGiftRespType:
				if live.SendGiftMessageHandler != nil {
					live.SendGiftMessageHandler(message.roomID, TransferSendGiftMessage(message.body))
				}
			case SpecialUserRespType:
				if live.SpecialUserMessageHandler != nil {
					live.SpecialUserMessageHandler(message.roomID, TransferSpecialUserMessage(message.body))
				}
			case SwitchBroadcastRespType:
				if live.SwitchBroadcastMessageHandler != nil {
					live.SwitchBroadcastMessageHandler(message.roomID, TransferSwitchBroadcastMessage(message.body))
				}
			case BroadcastRankRespType:
				if live.BroadcastRankMessageHandler != nil {
					live.BroadcastRankMessageHandler(message.roomID, TransferBroadcastRankMessage(message.body))
				}
			case SuperBarrageRespType:
				if live.SuperBarrageMessageHandler != nil {
					live.SuperBarrageMessageHandler(message.roomID, TransferSuperBarrageMessage(message.body))
				}
			case RoomGiftBroadcastRespType:
				if live.RoomGiftBarrageMessageHandler != nil {
					live.RoomGiftBarrageMessageHandler(message.roomID, TransferRoomGiftBroadcastMessage(message.body))
				}

			default:

			}

			break
		}

	}
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
		token, err := GenerateToken(room.aid, room.secret, currentTime)
		if err != nil {
			log.Panic(err)
		}
		room.token = token
		room.tokenTime = currentTime.Unix()
	}

	room.login(currentTime)

}

// 登录
func (room *liveRoom) login(currentTime time.Time) {
	auth := Md5(fmt.Sprintf("%s_%s_%d_%s", room.secret, room.aid, currentTime.Unix(), room.token))

	loginMessage := MsgToByte(map[string]string{"type": "loginreq", "roomid": strconv.Itoa(room.roomID), "aid": room.aid, "token": room.token, "time": strconv.FormatInt(currentTime.Unix(), 10), "auth": auth})

	fmt.Println(string(loginMessage))
	// 登录弹幕服务器
	if _, err := room.conn.Write(loginMessage); err != nil {
		log.Panic("login failed:", err)
	}

	room.auth = auth
}

// 入组
func (room *liveRoom) joinGroup() {

	joinGroupMessage := MsgToByte(map[string]string{
		"type":  "joingroup",
		"rid":   strconv.Itoa(room.roomID),
		"token": room.token,
		"time":  strconv.FormatInt(room.tokenTime, 10),
		"auth":  room.auth,
	})

	fmt.Println(string(joinGroupMessage))

	// 加入组
	if _, err := room.conn.Write(joinGroupMessage); err != nil {
		log.Panic("joinGroup failed:", err)
	}

}

// 心跳
func (room *liveRoom) heartBeat(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("heatbeat failed: %s", err)
		}
	}()
	time.Sleep(3 * time.Second)
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
				log.Println("尝试重新连接：", room.server, room.port)
				_, cancel := context.WithCancel(ctx)
				chReconSignal <- &liveRoom{
					roomID:    room.roomID,
					cancel:    cancel,
					server:    room.server,
					port:      room.port,
					token:     room.token,
					tokenTime: room.tokenTime,
					aid:       room.aid,
					secret:    room.secret,
					auth:      room.auth,
					reconnect: true,
				}
				break
			}
			log.Printf("heatbeat failed: %s", err.Error())
			errorCount++
		}
		errorCount = 0
		time.Sleep(30 * time.Second)
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
		n, err := room.conn.Read(headerBuffer)
		if err != nil {
			if err == io.EOF {
				continue
			}
			log.Println("read err:", err)
			continue
		}

		// 包体
		var messageBody = make([]byte, int(binary.LittleEndian.Uint32(headerBuffer[0:4]))-int(HeadLen+MsgTypeLen+KeepLen))
		n, err = room.conn.Read(messageBody)
		if err != nil {
			log.Println("read err:", err)
			continue
		}
		data := ByteToMsg(messageBody[:n])

		chSocketMessage <- &socketMessage{
			roomID: room.roomID,
			body:   data,
		}

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
