package media

import (
	"conference/util"
	"github.com/pion/webrtc/v2"
	"time"
)

var (
	webrtcEngine *WebRTCEngine
)

// 单例，饿汉模式，init
func init() {
	webrtcEngine = NewWebRTCEngine()
}

type WebRTCPeer struct {
	ID         string
	PC         *webrtc.PeerConnection
	VideoTrack *webrtc.Track
	AudioTrack *webrtc.Track
	stop       chan int
	pli        chan int
}

func NewWebRTCPeer(id string) *WebRTCPeer {
	return &WebRTCPeer{
		ID:   id,
		stop: make(chan int),
		pli:  make(chan int),
	}
}

func (p *WebRTCPeer) Stop() {
	close(p.stop)
	close(p.pli)
}

// 响应发送方，与发布方建立连接
func (p *WebRTCPeer) AnswerSender(offer webrtc.SessionDescription) (answer webrtc.SessionDescription, err error) {
	util.Infof("WebRTCPeer.AnswerSender")
	//创建接收
	return webrtcEngine.CreateReceiver(offer, &p.PC, &p.VideoTrack, &p.AudioTrack, p.stop, p.pli)
}

// 响应接收方，与订阅方建立连接
func (p *WebRTCPeer) AnswerReceiver(offer webrtc.SessionDescription, addVideoTrack **webrtc.Track, addAudioTrack **webrtc.Track) (answer webrtc.SessionDescription, err error) {
	util.Infof("WebRTCPeer.AnswerReceiver")
	//创建发送
	return webrtcEngine.CreateSender(offer, &p.PC, addVideoTrack, addAudioTrack, p.stop)
}

func (p *WebRTCPeer) SendPLI() {
	go func() {
		defer func() {
			//恢复
			if r := recover(); r != nil {
				util.Errorf("%v", r)
				return
			}
		}()
		ticker := time.NewTicker(time.Second)
		i := 0
		for {
			select {
			case <-ticker.C:
				p.pli <- 1
				if i > 3 {
					return
				}
				i++
			case <-p.stop:
				return
			}
		}
	}()
}
