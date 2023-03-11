package server

import (
	"conference/util"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

//服务配置
type SFUServerConfig struct {
	//IP
	Host string
	//端口
	Port int
	//Cert文件
	CertFile string
	//Key文件
	KeyFile string
	//Html根目录
	HTMLRoot string
	//WebSocket路径
	WebSocketPath string
}

//默认WebSocket服务配置
func DefaultConfig() SFUServerConfig {
	return SFUServerConfig{
		//IP
		Host: "0.0.0.0",
		//端口
		Port: 8000,
		//Html根目录
		HTMLRoot: "html",
		//WebSocket路径
		WebSocketPath: "/ws",
	}
}

//SFU服务
type SFUServer struct {
	//WebSocket绑定函数,由信令服务处理
	handleWebSocket func(ws *WebSocketConn, request *http.Request)
	//Websocket升级为长连接
	upgrader websocket.Upgrader
}

//实例化一个服务
func NewSFUServer(wsHandler func(ws *WebSocketConn, request *http.Request)) *SFUServer {
	//创建P2PServer对象
	var server = &SFUServer{
		//绑定WebSocket
		handleWebSocket: wsHandler,
	}
	//指定Websocket连接
	server.upgrader = websocket.Upgrader{
		//解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	//返回server
	return server
}

//WebSocket请求处理
func (server *SFUServer) handleWebSocketRequest(writer http.ResponseWriter, request *http.Request) {
	//返回头
	responseHeader := http.Header{}
	//responseHeader.Add("Sec-WebSocket-Protocol", "protoo")
	//升级为长连接
	socket, err := server.upgrader.Upgrade(writer, request, responseHeader)
	//输出错误日志
	if err != nil {
		util.Panicf("%v", err)
	}
	//实例化一个WebSocketConn对象
	wsTransport := NewWebSocketConn(socket)
	//处理具体的请求消息
	server.handleWebSocket(wsTransport, request)
	//WebSocketConn开始读取消息
	wsTransport.ReadMessage()
}

//绑定
func (server *SFUServer) Bind(cfg SFUServerConfig) {
	//WebSocket回调函数
	http.HandleFunc(cfg.WebSocketPath, server.handleWebSocketRequest)
	//Html绑定
	http.Handle("/", http.FileServer(http.Dir(cfg.HTMLRoot)))
	//输出日志
	util.Infof("SFU Server listening on: %s:%d", cfg.Host, cfg.Port)
	//启动并监听安全连接
	panic(http.ListenAndServeTLS(cfg.Host+":"+strconv.Itoa(cfg.Port), cfg.CertFile, cfg.KeyFile, nil))
}
