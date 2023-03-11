package room

import (
	"conference/media"
	"conference/server"
	"conference/util"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pion/webrtc/v2"
)

const (
	MethodJoin        = "join"
	MethodLeave       = "leave"
	MethodPublish     = "publish"
	MethodSubscribe   = "subscribe"
	MethodOnPublish   = "onPublish"
	MethodOnSubscribe = "onSubscribe"
	MethodOnUnpublish = "onUnpublish"
	MethodFindMark    = "findMark"
	MethodMark        = "mark"
	MethodFindEndMark = "findEndMark"
	MethodEndMark     = "endMark"
)

func (roomManager *RoomManager) HandleNewWebSocket(conn *server.WebSocketConn, request *http.Request) {
	util.Infof("On Open %v", request)
	//监听消息事件
	conn.On("message", func(message []byte) {
		//解析Json数据
		request, err := util.Unmarshal(string(message))

		if err != nil {
			util.Errorf("解析Json数据Unmarshal错误 %v", err)
			return
		}

		var data map[string]interface{} = nil

		tmp, found := request["data"]
		if !found {
			util.Errorf("没有发现数据")
			return
		}

		data = tmp.(map[string]interface{})

		roomId := data["roomId"].(string)
		util.Infof("房间Id:%v", roomId)

		room := roomManager.getRoom(roomId)

		if room == nil {
			room = roomManager.createRoom(roomId)
		}

		userId := data["userId"].(string)
		user := room.GetUser(userId)
		if user == nil {
			user = NewUser(userId, conn)
		}

		// roomid房间的userid用户，向sfu发送消息
		switch request["type"] {
		case MethodJoin:
			processJoin(user, data, roomManager)
			break
		case MethodPublish:
			processPublish(user, data, roomManager)
			break
		case MethodSubscribe:
			processSubscribe(user, data, roomManager)
			break
		case MethodLeave:
			processJoin(user, data, roomManager)
			break
		case MethodFindMark: // 增加sfu服务器端收到主讲人mark消息处理
			processFindMark(user, data, roomManager)
			break
		case MethodFindEndMark: // 增加sfu服务器端收到主讲人结束findEndMark消息处理
			processFindEndMark(user, data, roomManager)
			break
		default:
			{
				util.Warnf("未知的请求 %v", request)
			}
			break
		}

	})

	conn.On("close", func(code int, text string) {
		util.Infof("连接关闭 %v", conn)
		var userId string = ""
		var roomId string = ""

		for _, room := range roomManager.rooms {
			for _, user := range room.users {
				if user.conn == conn {
					userId = user.ID()
					roomId = room.ID
					break
				}
			}
		}

		if roomId == "" {
			util.Errorf("没有查找到退出的房间及用户")
			return
		}
		processLeave(roomId, userId, roomManager)
	})
}

func processLeave(roomId, userId string, roomManager *RoomManager) {

	room := roomManager.getRoom(roomId)
	if room == nil {
		return
	} else {

		onUnpublish := make(map[string]interface{})
		onUnpublish["pubid"] = userId

		for id, user := range room.users {
			if id != userId {
				user.sendMessage(MethodOnUnpublish, onUnpublish)
			}
		}
		room.delWebRTCPeer(userId, true)
		room.delWebRTCPeer(userId, false)
		room.DeleteUser(userId)
	}
}

func processJoin(user *User, message map[string]interface{}, roomManager *RoomManager) {
	roomId := message["roomId"]
	if roomId == nil {
		return
	}

	room := roomManager.getRoom(roomId.(string))
	if room == nil {
		room = roomManager.createRoom(roomId.(string))
	}

	room.AddUser(user)
	onPublish := make(map[string]interface{})

	room.pubPeerLock.RLock()
	defer room.pubPeerLock.RUnlock()
	//找到当前房间的所有发布者
	for peerId, _ := range room.pubPeers {
		if peerId != user.ID() {
			onPublish["pubid"] = peerId
			onPublish["userId"] = peerId
			room.GetUser(user.ID()).sendMessage(MethodOnPublish, onPublish)
		}
	}

	onJoinData := make(map[string]interface{})
	onJoinData["status"] = "success"
	user.sendMessage("onJoinRoom", onJoinData)
	log.Print("onJoinRoom")
}

func processPublish(user *User, message map[string]interface{}, roomManager *RoomManager) {
	if message["jsep"] == nil {
		log.Print("jsep...")
		return
	}
	j := message["jsep"].(map[string]interface{})
	if j["sdp"] == nil {
		log.Print("sdp...")
		return
	}

	roomId := message["roomId"]
	r := roomManager.getRoom(roomId.(string))
	if r == nil {
		log.Print("room...")
		return
	}
	r.addWebRTCPeer(user.ID(), true)
	jsep := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  j["sdp"].(string),
	}
	answer, err := r.answer(user.ID(), "", jsep, true)
	if err != nil {
		log.Print("创建Answer失败")
		return
	}

	resp := make(map[string]interface{})
	resp["jsep"] = answer
	resp["userId"] = user.ID()
	respByte, err := json.Marshal(resp)
	if err != nil {
		return
	}
	respStr := string(respByte)
	if respStr != "" {
		//返回给自己jsep
		user.sendMessage(MethodOnPublish, resp)

		onPublish := make(map[string]interface{})
		onPublish["pubid"] = user.ID()
		//发送给房间其他人jsep
		r.sendMessage(user, MethodOnPublish, resp)
		return
	}

}

func processSubscribe(user *User, message map[string]interface{}, roomManager *RoomManager) {
	if message["jsep"] == nil {
		log.Print("jsep...")
		return
	}
	j := message["jsep"].(map[string]interface{})
	if j["sdp"] == nil {
		log.Print("sdp...")
		return
	}

	roomId := message["roomId"]
	r := roomManager.getRoom(roomId.(string))
	if r == nil {
		log.Print("room...")
		return
	}

	r.addWebRTCPeer(user.ID(), false)
	jsep := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  j["sdp"].(string),
	}
	answer, err := r.answer(user.ID(), message["pubid"].(string), jsep, false)
	if err != nil {
		log.Print("创建Answer失败")
		return
	}

	resp := make(map[string]interface{})
	resp["jsep"] = answer
	resp["userId"] = user.ID()
	resp["pubid"] = message["pubid"]

	respByte, err := json.Marshal(resp)
	if err != nil {
		log.Print(err.Error())
		return
	}
	r.sendPLI(user.ID())
	respStr := string(respByte)

	if respStr != "" {
		//返回给自己jsep
		user.sendMessage(MethodOnSubscribe, resp)
		log.Printf("Subscribe返回给自己的Id:%s", user.ID())
		return
	}

}

// 读取主讲人发来的findMark信令，然后给房间内其他用户发送Mark消息，进行标识主讲人
func processFindMark(user *User, message map[string]interface{}, roomManager *RoomManager) {
	roomId := message["roomId"]
	r := roomManager.getRoom(roomId.(string))
	if r == nil {
		log.Print("room...")
		return
	}

	markSignal := make(map[string]interface{})
	markSignal["userId"] = user.ID()
	markSignal["pubid"] = message["userId"]

	// 给其他用户发送消息改变颜色，user.send是也给自己发送消息改变颜色
	r.sendMessage(user, MethodMark, markSignal)
	// user.sendMessage(MethodMark, markSignal)

}

// 读取主讲人发言结束发来的endMark信令，然后给房间内其他用户发送Mark消息，结束标记主讲人
func processFindEndMark(user *User, message map[string]interface{}, roomManager *RoomManager) {
	roomId := message["roomId"]
	r := roomManager.getRoom(roomId.(string))
	if r == nil {
		log.Print("room...")
		return
	}

	markSignal := make(map[string]interface{})
	markSignal["userId"] = user.ID()
	markSignal["pubid"] = message["userId"]

	// 给其他用户发送消息变回原来颜色，user.send是也给自己发送消息改变颜色
	r.sendMessage(user, MethodEndMark, markSignal)
	// user.sendMessage(MethodMark, markSignal)

}

type RoomManager struct {
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	var roomManager = &RoomManager{
		rooms: make(map[string]*Room),
	}
	return roomManager
}

type Room struct {
	users map[string]*User

	ID string

	pubPeers    map[string]*media.WebRTCPeer
	subPeers    map[string]*media.WebRTCPeer
	pubPeerLock sync.RWMutex
	subPeerLock sync.RWMutex
}

func NewRoom(id string) *Room {
	var room = &Room{
		users:    make(map[string]*User),
		pubPeers: make(map[string]*media.WebRTCPeer),
		subPeers: make(map[string]*media.WebRTCPeer),
		ID:       id,
	}
	return room
}

func (roomManager *RoomManager) getRoom(id string) *Room {
	return roomManager.rooms[id]
}

func (roomManager *RoomManager) createRoom(id string) *Room {
	roomManager.rooms[id] = NewRoom(id)
	return roomManager.rooms[id]
}

func (roomManager *RoomManager) deleteRoom(id string) {
	delete(roomManager.rooms, id)
}

func (room *Room) AddUser(newUser *User) {
	room.users[newUser.ID()] = newUser
}

func (room *Room) GetUser(userId string) *User {

	if user, ok := room.users[userId]; ok {
		return user
	}
	return nil
}

func (room *Room) DeleteUser(userId string) {
	delete(room.users, userId)
}

func (room *Room) getWebRTCPeer(id string, sender bool) *media.WebRTCPeer {
	if sender {
		room.pubPeerLock.Lock()
		defer room.pubPeerLock.Unlock()
		return room.pubPeers[id]
	} else {
		room.subPeerLock.Lock()
		defer room.subPeerLock.Unlock()
		return room.subPeers[id]
	}
}

func (r *Room) delWebRTCPeer(id string, sender bool) {
	if sender {
		r.pubPeerLock.Lock()
		defer r.pubPeerLock.Unlock()
		if r.pubPeers[id] != nil {
			if r.pubPeers[id].PC != nil {
				r.pubPeers[id].PC.Close()
			}
			r.pubPeers[id].Stop()
		}
		delete(r.pubPeers, id)
	} else {
		r.subPeerLock.Lock()
		defer r.subPeerLock.Unlock()
		if r.subPeers[id] != nil {
			if r.subPeers[id].PC != nil {
				r.subPeers[id].PC.Close()
			}
			r.subPeers[id].Stop()
		}
		delete(r.subPeers, id)
	}

}

func (room *Room) addWebRTCPeer(id string, sender bool) {
	if sender {
		room.pubPeerLock.Lock()
		defer room.pubPeerLock.Unlock()
		if room.pubPeers[id] != nil {
			room.pubPeers[id].Stop()
		}
		room.pubPeers[id] = media.NewWebRTCPeer(id)
	} else {
		room.subPeerLock.Lock()
		defer room.subPeerLock.Unlock()
		if room.subPeers[id] != nil {
			room.subPeers[id].Stop()
		}
		room.subPeers[id] = media.NewWebRTCPeer(id)
	}
}

//关键侦丢包重传
func (r *Room) sendPLI(skipID string) {
	log.Print("Room.sendPLI")
	r.pubPeerLock.RLock()
	defer r.pubPeerLock.RUnlock()
	for k, v := range r.pubPeers {
		if k != skipID {
			v.SendPLI()
		}
	}
}

func (room *Room) sendMessage(from *User, msgType string, data map[string]interface{}) {

	var message map[string]interface{} = nil

	message = map[string]interface{}{
		"type": msgType,
		"data": data,
	}

	for id, user := range room.users {
		if id != from.ID() {
			user.conn.Send(util.Marshal(message))
		}
	}

}

func (r *Room) answer(id string, pubid string, offer webrtc.SessionDescription, sender bool) (webrtc.SessionDescription, error) {

	p := r.getWebRTCPeer(id, sender)

	var err error
	var answer webrtc.SessionDescription
	if sender {
		answer, err = p.AnswerSender(offer)
	} else {
		r.pubPeerLock.RLock()

		pub := r.pubPeers[pubid]
		r.pubPeerLock.RUnlock()
		ticker := time.NewTicker(time.Millisecond * 2000)
		for {
			select {
			case <-ticker.C:
				goto ENDWAIT
			default:
				if pub.VideoTrack == nil || pub.AudioTrack == nil {
					time.Sleep(time.Millisecond * 100)
				} else {
					goto ENDWAIT
				}
			}
		}

	ENDWAIT:
		answer, err = p.AnswerReceiver(offer, &pub.VideoTrack, &pub.AudioTrack)

	}
	return answer, err

}

func (r *Room) Close() {
	r.pubPeerLock.Lock()
	defer r.pubPeerLock.Unlock()
	for _, v := range r.pubPeers {
		if v != nil {
			v.Stop()
			if v.PC != nil {
				v.PC.Close()
			}
		}
	}
}
