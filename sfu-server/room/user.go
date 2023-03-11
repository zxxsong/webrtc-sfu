package room

import (
	"conference/server"
	"conference/util"

	"github.com/chuckpreslar/emission"
)

type User struct {
	emission.Emitter
	id   string
	conn *server.WebSocketConn
}

func NewUser(id string, conn *server.WebSocketConn) *User {
	var user User
	user.Emitter = *emission.NewEmitter()
	user.id = id
	user.conn = conn
	return &user
}

func (user *User) Close() {
	user.conn.Close()
}

func (user *User) ID() string {
	return user.id
}

func (user *User) sendMessage(msgType string, data map[string]interface{}) {

	var message map[string]interface{} = nil

	message = map[string]interface{}{
		"type": msgType,
		"data": data,
	}

	user.conn.Send(util.Marshal(message))
}
