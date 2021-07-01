package src

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
)

type AuthInfo struct {
	Label                string                 `json:"label"`
	AvatarUrl            string                 `json:"avatar_url"`
	WebrtcConnectionInfo map[string]interface{} `json:"webrtc_connection_info"`
}

type PeerMessage struct {
	Event   string                 `json:"event"`
	Uid     string                 `json:"uid"`
	Message map[string]interface{} `json:"message"`
}

type UserAddedEvent struct {
	AuthInfo

	Event string `json:"event"`
	Uid   string `json:"uid"`
}

type UserRemovedEvent struct {
	Event string `json:"event"`
	Uid   string `json:"uid"`
}

type RoomInfo struct {
	RoomId string
	Users  map[string]bool
}

type UserInfo struct {
	authInfo *AuthInfo
	ws       *websocket.Conn
	roomInfo *RoomInfo
}

type RoomService struct {
	userInfos map[string]*UserInfo
	roomInfos map[string]*RoomInfo
}

var mu sync.Mutex
var roomService RoomService

func init() {
	// TODO: use mutex or channel. to avoid Concurrent issue
	roomService.userInfos = make(map[string]*UserInfo)
	roomService.roomInfos = make(map[string]*RoomInfo)
}

func (r *RoomService) getOrCreateRoom(roomId string) *RoomInfo {
	if _, ok := r.roomInfos[roomId]; !ok {
		r.roomInfos[roomId] = &RoomInfo{
			RoomId: roomId,
			Users:  make(map[string]bool),
		}
	}
	return r.roomInfos[roomId]
}

func (r *RoomService) broadcastUserAdded(roomInfo *RoomInfo, newUserId string) {
	var userAddedEvent = UserAddedEvent{
		Event:    "user_added",
		Uid:      newUserId,
		AuthInfo: *r.userInfos[newUserId].authInfo,
	}
	for uid := range roomInfo.Users {
		if uid == newUserId {
			continue
		}
		r.userInfos[uid].ws.WriteJSON(userAddedEvent)
	}
}

func (r *RoomService) broadcastUserRemoved(roomInfo *RoomInfo, removedUserId string) {
	userRemovedEvent := UserRemovedEvent{
		Event: "user_removed",
		Uid:   removedUserId,
	}
	for uid := range roomInfo.Users {
		if uid == removedUserId {
			continue
		}
		r.userInfos[uid].ws.WriteJSON(userRemovedEvent)
	}
}

func (r *RoomService) listUsersWhenAdded(roomInfo *RoomInfo, newUserId string) {
	for uid := range roomInfo.Users {
		if uid == newUserId {
			continue
		}
		var userAddedEvent = UserAddedEvent{
			Event:    "user_added",
			Uid:      uid,
			AuthInfo: *r.userInfos[uid].authInfo,
		}
		r.userInfos[newUserId].ws.WriteJSON(userAddedEvent)
	}
}

func (r *RoomService) AddUser(roomId, uid string, ws *websocket.Conn, authInfo *AuthInfo) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := r.userInfos[uid]; ok {
		return fmt.Errorf("user uid duplicated")
	}

	roomInfo := r.getOrCreateRoom(roomId)
	roomInfo.Users[uid] = true

	userInfo := UserInfo{
		authInfo: authInfo,
		ws:       ws,
		roomInfo: roomInfo,
	}
	r.userInfos[uid] = &userInfo

	r.broadcastUserAdded(roomInfo, uid)
	r.listUsersWhenAdded(roomInfo, uid)
	return nil
}

func (r *RoomService) RemoveUser(uid string) {
	mu.Lock()
	defer mu.Unlock()

	roomInfo := r.userInfos[uid].roomInfo

	r.broadcastUserRemoved(roomInfo, uid)
	delete(roomInfo.Users, uid)

	if len(roomInfo.Users) == 0 {
		delete(r.roomInfos, roomInfo.RoomId)
	}
	delete(r.userInfos, uid)
}

func (r *RoomService) SendMessage(from, to string, message map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	targetUser, ok := r.userInfos[to]
	if !ok {
		return
	}
	peerMessage := PeerMessage{
		Event:   "peer_message",
		Uid:     from,
		Message: message,
	}
	targetUser.ws.WriteJSON(peerMessage)
}
