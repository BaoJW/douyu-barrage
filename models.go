package douyulive

import (
	"context"
	"net"
	"sync"
)

// Live 直播间
type Live struct {
	Debug                     bool                              // 是否显示日志
	AnalysisRoutineNum        int                               // 消息分析协程数量，默认为1，为1可以保证通知顺序与接收到消息顺序相同
	LoginRespMessageHandler   func(int, *LoginRespMessageModel) // 登录响应消息handler
	BarrageMessageHandler     func(int, *BarrageMessageModel)   // 弹幕消息handler
	StormMessageHandler       func(int, *StormMessage)          // 领取在线鱼丸暴击消息handler
	SendGiftMessageHandler    func(int, *SendGiftMessage)       // 赠送礼物消息handler
	SpecialUserMessageHandler func(int, *SpecialUserMessage)    // 用户进房通知消息handler
	wg                        sync.WaitGroup
	ctx                       context.Context

	chSocketMessage chan *socketMessage

	room map[int]*liveRoom // 直播间
}

type socketMessage struct {
	roomID int // 房间ID
	body   map[string]string
}

type liveRoom struct {
	roomID             int // 房间ID
	cancel             context.CancelFunc
	server             string // 地址
	port               int    // 端口
	hostServerList     []*hostServerList
	currentServerIndex int
	token              string // key
	tokenTime          int64  // key
	conn               *net.TCPConn
	aid                string
	secret             string
	auth               string
}

type hostServerList struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	WssPort int    `json:"wss_port"`
	WsPort  int    `json:"ws_port"`
}

// 登录响应消息模型
type LoginRespMessageModel struct {
	Type          string `json:"type"`       // 表示为“登录”消息，固定为 loginres
	UserID        int64  `json:"userid"`     // 用户 ID
	RoomGroup     int64  `json:"roomgroup"`  // 房间权限组
	PlatformGroup int64  `json:"pg"`         // 平台权限组
	SessioniID    int64  `json:"sessionid"`  // 会话ID
	UserName      string `json:"username"`   // 用户名
	NickName      string `json:"nickname"`   // 用户昵称
	LiveStat      int64  `json:"live_stat"`  // 直播状态
	IsIllegal     int64  `json:"is_illegal"` // 是否违规
	IllContent    string `json:"ill_ct"`     // 违规提醒内容
	IllTimestamp  int64  `json:"ill_ts"`     // 违规提醒开始时间戳
	Now           int64  `json:"now"`        // 系统当前时间
	Ps            int64  `json:"ps"`         // 手机绑定标示
	Es            int64  `json:"es"`         // 邮箱绑定标示
	It            int64  `json:"it"`         // 认证类型
	Its           int64  `json:"its"`        // 认证状态
	Npv           int64  `json:"npv"`        // 是否需要手机验证
	BestDlev      int64  `json:"best_dlev"`  // 最高酬勤等级
	CurLev        int64  `json:"cur_lev"`    // 酬勤等级
	Nrc           int64  `json:"nrc"`        // 观看房间需要的条件
	Ih            int64  `json:"ih"`         // 是否进房隐身
	SID           int64  `json:"sid"`        // 服务 id
	Sahf          int64  `json:"sahf"`       // 扩展字段，一般不使用，可忽略
}

// BarrageMessageModel 弹幕消息模型
type BarrageMessageModel struct {
	Type                string    `json:"type"`   // 表示为“弹幕”消息，固定为 chatmsg
	GroupID             int64     `json:"gid"`    // 弹幕组id
	RoomID              int64     `json:"rid"`    // 房间id
	UID                 int64     `json:"uid"`    // 发送者uid
	NickName            string    `json:"nn"`     // 发送者昵称
	Txt                 string    `json:"txt"`    // 弹幕文本内容
	CID                 int64     `json:"cid"`    // 弹幕唯一ID
	Level               int64     `json:"level"`  // 用户等级
	GiftTitle           int64     `json:"gt"`     // 礼物头衔：默认值 0（表示没有头衔）
	Color               int64     `json:"col"`    // 颜色：默认值 0（表示默认颜色弹幕）
	ClientType          int64     `json:"ct"`     // 客户端类型：默认值 0
	RoomGroup           int64     `json:"rg"`     // 房间权限组：默认值 1（表示普通权限用户）
	PlatformGroup       int64     `json:"pg"`     // 平台权限组：默认值 1（表示普通权限用户）
	DiligentLevel       int64     `json:"dlv"`    // 酬勤等级：默认值 0（表示没有酬勤）
	DiligentCount       int64     `json:"dc"`     // 酬勤数量：默认值 0（表示没有酬勤数量）
	BestDiligentLevel   int64     `json:"bdlv"`   // 最高酬勤等级：默认值 0（表示全站都没有酬勤）
	ChatMsgType         int64     `json:"cmt"`    // 弹幕具体类型: 默认值 0（普通弹幕）
	Sahf                int64     `json:"sahf"`   // 扩展字段，一般不使用，可忽略
	Ic                  string    `json:"ic"`     // 用户头像
	NobleLevel          int64     `json:"nl"`     // 贵族等级
	NobleChat           int64     `json:"nc"`     // 贵族弹幕标识,0-非贵族弹幕,1-贵族弹幕,默认值 0
	GatewayTimestampIn  int64     `json:"gatin"`  // 进入网关服务时间戳
	GatewayTimestampOut int64     `json:"gatout"` // 离开网关服务时间戳
	ChtIn               int64     `json:"chtin"`  // 进入房间服务时间戳
	ChtOut              int64     `json:"chtout"` // 离开房间服务时间戳
	Repin               int64     `json:"repin"`  // 进入发送服务时间戳
	Repout              int64     `json:"repout"` // 离开发送服务时间戳
	BadgeNickName       string    `json:"bnn"`    // 徽章昵称
	BadgeLevel          int64     `json:"bl"`     // 徽章等级
	BadgeRoomID         int64     `json:"brid"`   // 徽章房间 id
	Hc                  int64     `json:"hc"`     // 徽章信息校验码
	AnchorLevel         int64     `json:"ol"`     // 主播等级
	Reserve             int64     `json:"rev"`    // 是否反向弹幕标记: 0-普通弹幕，1-反向弹幕, 默认值 0
	HighLight           int64     `json:"hl"`     // 否高亮弹幕标记: 0-普通，1-高亮, 默认值 0
	Ifs                 int64     `json:"ifs"`    // 是否粉丝弹幕标记: 0-非粉丝弹幕，1-粉丝弹幕, 默认值 0
	P2P                 int64     `json:"p2p"`    // 服务功能字段
	El                  *ElDetail `json:"el"`     // 用户获得的连击特效
}

type ElDetail struct {
	EID   int64 `json:"eid"` // 特效 id
	EType int64 `json:"etp"` // 特效类型
	Sc    int64 `json:"sc"`  // 特效次数
	Ef    int64 `json:"ef"`  // 特效标志
}

// 领取在线鱼丸暴击消息 在线领取鱼丸时，若出现暴击，服务则发送领取暴击消息到客户端。
type StormMessage struct {
	Type          string `json:"type"`  // 表示为“领取在线鱼丸”消息，固定为 onlinegift
	RoomID        int64  `json:"rid"`   // 房间ID
	UserID        int64  `json:"uid"`   // 用户ID
	GroupID       int64  `json:"gid"`   // 弹幕分组ID
	Sil           int64  `json:"sil"`   // 鱼丸数
	If            int64  `json:"if"`    // 领取鱼丸的等级
	Ct            int64  `json:"ct"`    // 客户端类型标识
	NickName      string `json:"nn"`    // 用户昵称
	Ur            int64  `json:"ur"`    // 鱼丸之刃倍率
	Level         int64  `json:"level"` // 用户等级
	BroadcastType int64  `json:"btype"` // 广播类型
}

//  赠送礼物消息 用户在房间赠送礼物时，服务端发送此消息给客户端
type SendGiftMessage struct {
	Type     string `json:"type"`  // 表示为“赠送礼物”消息，固定为 dgb
	RoomID   int64  `json:"rid"`   // 房间ID
	GroupID  int64  `json:"gid"`   // 弹幕分组ID
	GiftID   int64  `json:"gfid"`  // 礼物 id
	Gs       int64  `json:"gs"`    // 礼物显示样式
	UserID   int64  `json:"uid"`   // 用户ID
	NickName string `json:"nn"`    // 用户昵称
	Bg       int64  `json:"bg"`    // 大礼物标识：默认值为 0（表示是小礼物）
	Ic       int64  `json:"ic"`    // 用户头像
	EID      int64  `json:"eid"`   // 礼物关联的特效 id
	Level    int64  `json:"level"` // 用户等级
	Dw       int64  `json:"dw"`    // 主播体重
	GfCount  int64  `json:"gfcnt"` // 礼物个数：默认值 1（表示 1 个礼物）
	Hits     int64  `json:"hits"`  // 礼物连击次数：默认值 1（表示 1 连击）
	Dlv      int64  `json:"dlv"`   // 酬勤头衔：默认值 0（表示没有酬勤）
	Dc       int64  `json:"dc"`    // 酬勤个数：默认值 0（表示没有酬勤数量）
	Bdl      int64  `json:"bdl"`   // 全站最高酬勤等级：默认值 0（表示全站都没有酬勤）
	Rg       int64  `json:"rg"`    // 房间身份组：默认值 1（表示普通权限用户）
	Pg       int64  `json:"pg"`    // 平台身份组：默认值 1（表示普通权限用户）
	RpID     int64  `json:"rpid"`  // 扩展字段 id
	RpIDn    int64  `json:"rpidn"` // 扩展字段 id
	Slt      int64  `json:"slt"`   // 扩展字段，一般不使用
	Elt      int64  `json:"elt"`   // 扩展字段，一般不使用
	Nl       int64  `json:"nl"`    // 贵族等级：默认值 0（表示不是贵族）
	Sahf     int64  `json:"sahf"`  // 扩展字段，一般不使用，可忽略
	BNN      string `json:"bnn"`   // 徽章昵称
	BL       int64  `json:"bl"`    // 徽章等级
	Brid     int64  `json:"brid"`  // 徽章房间 id
	Hc       int64  `json:"hc"`    // 徽章信息校验码
	Fc       int64  `json:"fc"`    // 攻击道具的攻击力
}

// 用户进房通知消息 具有特殊属性的用户进入直播间时，服务端发送此消息至客户端
type SpecialUserMessage struct {
	Type     string    `json:"type"`  // 表示为“用户进房通知”消息，固定为 uenter
	RoomID   int64     `json:"rid"`   // 房间ID
	GroupID  int64     `json:"gid"`   // 弹幕分组ID
	NickName string    `json:"nn"`    // 用户昵称
	Str      int64     `json:"str"`   // 战斗力
	Level    int64     `json:"level"` // 新用户等级
	Gt       int64     `json:"gt"`    // 礼物头衔：默认值 0（表示没有头衔）
	Rg       int64     `json:"rg"`    // 房间权限组：默认值 1（表示普通权限用户）
	Pg       int64     `json:"pg"`    // 平台身份组：默认值 1（表示普通权限用户）
	Dlv      int64     `json:"dlv"`   // 酬勤等级：默认值 0（表示没有酬勤）
	Dc       int64     `json:"dc"`    // 酬勤数量：默认值 0（表示没有酬勤数量）
	Bdlv     int64     `json:"bdlv"`  // 最高酬勤等级：默认值 0
	Ic       int64     `json:"ic"`    // 用户头像
	Nl       int64     `json:"nl"`    // 贵族等级
	CeID     int64     `json:"ceid"`  // 扩展功能字段 id
	Crw      int64     `json:"crw"`   // 用户栏目上周排名
	Ol       int64     `json:"ol"`    // 主播等级
	El       *ElDetail `json:"el"`
	Sahf     int64     `json:"sahf"` // 扩展字段，一般不使用，可忽略
	Wgei     int64     `json:"wgei"` // 页游欢迎特效 id

}

func TransferLoginRespMessage(data map[string]string) *LoginRespMessageModel {
	return &LoginRespMessageModel{
		Type:          data["type"],
		UserID:        StrToInt64(data["userid"]),
		RoomGroup:     StrToInt64(data["roomgroup"]),
		PlatformGroup: StrToInt64(data["pg"]),
		SessioniID:    StrToInt64(data["sessionid"]),
		UserName:      data["username"],
		NickName:      data["nickname"],
		LiveStat:      StrToInt64(data["live_stat"]),
		IsIllegal:     StrToInt64(data["is_illegal"]),
		IllContent:    data["ill_ct"],
		IllTimestamp:  StrToInt64(data["ill_ts"]),
		Now:           StrToInt64(data["now"]),
		Ps:            StrToInt64(data["ps"]),
		Es:            StrToInt64(data["es"]),
		It:            StrToInt64(data["it"]),
		Its:           StrToInt64(data["its"]),
		Npv:           StrToInt64(data["npv"]),
		BestDlev:      StrToInt64(data["best_dlev"]),
		CurLev:        StrToInt64(data["cur_lev"]),
		Nrc:           StrToInt64(data["nrc"]),
		Ih:            StrToInt64(data["ih"]),
		SID:           StrToInt64(data["sid"]),
		Sahf:          StrToInt64(data["sahf"]),
	}
}

func TransferBarrageMessage(data map[string]string) *BarrageMessageModel {
	return &BarrageMessageModel{
		Type:                data["type"],
		GroupID:             StrToInt64(data["gid"]),
		RoomID:              StrToInt64(data["rid"]),
		UID:                 StrToInt64(data["uid"]),
		NickName:            data["nn"],
		Txt:                 data["txt"],
		CID:                 StrToInt64(data["cid"]),
		Level:               StrToInt64(data["level"]),
		GiftTitle:           StrToInt64(data["gt"]),
		Color:               StrToInt64(data["col"]),
		ClientType:          StrToInt64(data["ct"]),
		RoomGroup:           StrToInt64(data["rg"]),
		PlatformGroup:       StrToInt64(data["pg"]),
		DiligentLevel:       StrToInt64(data["dlv"]),
		DiligentCount:       StrToInt64(data["dc"]),
		BestDiligentLevel:   StrToInt64(data["bdlv"]),
		ChatMsgType:         StrToInt64(data["cmt"]),
		Sahf:                StrToInt64(data["sahf"]),
		Ic:                  data["ic"],
		NobleLevel:          StrToInt64(data["nl"]),
		NobleChat:           StrToInt64(data["nc"]),
		GatewayTimestampIn:  StrToInt64(data["gatin"]),
		GatewayTimestampOut: StrToInt64(data["gatout"]),
		ChtIn:               StrToInt64(data["chtin"]),
		ChtOut:              StrToInt64(data["chtout"]),
		Repin:               StrToInt64(data["repin"]),
		Repout:              StrToInt64(data["repout"]),
		BadgeNickName:       data["bnn"],
		BadgeLevel:          StrToInt64(data["bl"]),
		BadgeRoomID:         StrToInt64(data["brid"]),
		Hc:                  StrToInt64(data["hc"]),
		AnchorLevel:         StrToInt64(data["ol"]),
		Reserve:             StrToInt64(data["rev"]),
		HighLight:           StrToInt64(data["hl"]),
		Ifs:                 StrToInt64(data["ifs"]),
		P2P:                 StrToInt64(data["p2p"]),
		El: &ElDetail{
			EID:   StrToInt64(data["eid"]),
			EType: StrToInt64(data["etp"]),
			Sc:    StrToInt64(data["sc"]),
			Ef:    StrToInt64(data["ef"]),
		},
	}
}

func TransferStormMessage(data map[string]string) *StormMessage {
	return &StormMessage{
		Type:          data["type"],
		RoomID:        StrToInt64(data["rid"]),
		UserID:        StrToInt64(data["uid"]),
		GroupID:       StrToInt64(data["gid"]),
		Sil:           StrToInt64(data["sil"]),
		If:            StrToInt64(data["if"]),
		Ct:            StrToInt64(data["ct"]),
		NickName:      data["nn"],
		Ur:            StrToInt64(data["ur"]),
		Level:         StrToInt64(data["level"]),
		BroadcastType: StrToInt64(data["btype"]),
	}
}

func TransferSendGiftMessage(data map[string]string) *SendGiftMessage {
	return &SendGiftMessage{
		Type:     data["type"],
		RoomID:   StrToInt64(data["rid"]),
		GroupID:  StrToInt64(data["gid"]),
		GiftID:   StrToInt64(data["gfid"]),
		Gs:       StrToInt64(data["gs"]),
		UserID:   StrToInt64(data["uid"]),
		NickName: data["nn"],
		Bg:       StrToInt64(data["bg"]),
		Ic:       StrToInt64(data["ic"]),
		EID:      StrToInt64(data["eid"]),
		Level:    StrToInt64(data["level"]),
		Dw:       StrToInt64(data["dw"]),
		GfCount:  StrToInt64(data["gfcnt"]),
		Hits:     StrToInt64(data["hits"]),
		Dlv:      StrToInt64(data["dlv"]),
		Dc:       StrToInt64(data["dc"]),
		Bdl:      StrToInt64(data["bdl"]),
		Rg:       StrToInt64(data["rg"]),
		Pg:       StrToInt64(data["pg"]),
		RpID:     StrToInt64(data["rpid"]),
		RpIDn:    StrToInt64(data["rpidn"]),
		Slt:      StrToInt64(data["slt"]),
		Elt:      StrToInt64(data["elt"]),
		Nl:       StrToInt64(data["nl"]),
		Sahf:     StrToInt64(data["sahf"]),
		BNN:      data["bnn"],
		BL:       StrToInt64(data["bl"]),
		Brid:     StrToInt64(data["brid"]),
		Hc:       StrToInt64(data["hc"]),
		Fc:       StrToInt64(data["fc"]),
	}
}

func TransferSpecialUserMessage(data map[string]string) *SpecialUserMessage {
	return &SpecialUserMessage{
		Type:     data["type"],
		RoomID:   StrToInt64(data["rid"]),
		GroupID:  StrToInt64(data["gid"]),
		NickName: data["nn"],
		Str:      StrToInt64(data["str"]),
		Level:    StrToInt64(data["level"]),
		Gt:       StrToInt64(data["gt"]),
		Rg:       StrToInt64(data["rg"]),
		Pg:       StrToInt64(data["pg"]),
		Dlv:      StrToInt64(data["dlv"]),
		Dc:       StrToInt64(data["dc"]),
		Bdlv:     StrToInt64(data["bdlv"]),
		Ic:       StrToInt64(data["ic"]),
		Nl:       StrToInt64(data["nl"]),
		CeID:     StrToInt64(data["ceid"]),
		Crw:      StrToInt64(data["crw"]),
		Ol:       StrToInt64(data["ol"]),
		El: &ElDetail{
			EID:   StrToInt64(data["eid"]),
			EType: StrToInt64(data["etp"]),
			Sc:    StrToInt64(data["sc"]),
			Ef:    StrToInt64(data["ef"]),
		},
		Sahf: StrToInt64(data["sahf"]),
		Wgei: StrToInt64(data["wgei"]),
	}
}
