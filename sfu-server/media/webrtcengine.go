package media

import (
	"conference/util"
	"io"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
)

var defaultPeerCfg = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.stunprotocol.org:3478"},
		},
	},
}

const (
	//一个媒体传送单元是1400 分成7个包 每侦所需要的RTP包的个数
	averageRtpPacketsPerFrame = 7
)

type WebRTCEngine struct {
	cfg webrtc.Configuration

	mediaEngine webrtc.MediaEngine

	api *webrtc.API
}

func NewWebRTCEngine() *WebRTCEngine {
	urls := []string{} //conf.SFU.Ices//[]string{"stun:stun.stunprotocol.org:3478"};//conf.SFU.Ices

	w := &WebRTCEngine{
		mediaEngine: webrtc.MediaEngine{},
		cfg: webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
			ICEServers: []webrtc.ICEServer{
				{
					URLs: urls,
				},
			},
		},
	}

	w.mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	w.mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	w.api = webrtc.NewAPI(webrtc.WithMediaEngine(w.mediaEngine))
	return w
}

// 创建发送数据对象 朝接收者发送数据
func (s WebRTCEngine) CreateSender(offer webrtc.SessionDescription, pc **webrtc.PeerConnection, addVideoTrack, addAudioTrack **webrtc.Track, stop chan int) (answer webrtc.SessionDescription, err error) {

	*pc, err = s.api.NewPeerConnection(s.cfg)
	util.Infof("WebRTCEngine.CreateSender pc=%p", *pc)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if *addVideoTrack != nil && *addAudioTrack != nil {
		(*pc).AddTrack(*addVideoTrack)
		(*pc).AddTrack(*addAudioTrack)
		err = (*pc).SetRemoteDescription(offer)
		if err != nil {
			return webrtc.SessionDescription{}, err
		}
	}

	//创建应答Answer
	answer, err = (*pc).CreateAnswer(nil)
	//设置本地SDP
	err = (*pc).SetLocalDescription(answer)
	util.Infof("WebRTCEngine.CreateReceiver ok")
	return answer, err

}

// 创建接收者对象
func (s WebRTCEngine) CreateReceiver(offer webrtc.SessionDescription, pc **webrtc.PeerConnection, videoTrack, audioTrack **webrtc.Track, stop chan int, pli chan int) (answer webrtc.SessionDescription, err error) {

	*pc, err = s.api.NewPeerConnection(s.cfg)
	util.Infof("WebRTCEngine.CreateReceiver pc=%p", *pc)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	_, err = (*pc).AddTransceiver(webrtc.RTPCodecTypeVideo)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	_, err = (*pc).AddTransceiver(webrtc.RTPCodecTypeAudio)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	// 监听OnTrack事件
	// OnTrack sets an event handler which is called when remote track arrives from a remote peer.
	(*pc).OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {

		//视频处理
		if remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeVP8 ||
			remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeVP9 ||
			remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeH264 {
			//根据remoteTrack创建一个VideoTrack赋值给videoTrack
			*videoTrack, err = (*pc).NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), "video", remoteTrack.Label())

			go func() {
				for {
					select {
					case <-pli:
						//PictureLossIndication 关键帧丢包重传,参考rfc4585  SSRC同步源标识符
						(*pc).WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}})
					case <-stop:
						return
					}
				}
			}()
			//rtp解包
			var pkt rtp.Depacketizer
			//判断视频编码
			if remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeVP8 {
				//使用VP8编码
				pkt = &codecs.VP8Packet{}
			} else if remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeVP9 {
				util.Errorf("TODO codecs.VP9Packet")
			} else if remoteTrack.PayloadType() == webrtc.DefaultPayloadTypeH264 {
				util.Errorf("TODO codecs.H264Packet")
			}

			// SampleBuilder contains all packets
			// maxLate determines how long we should wait until we get a valid Sample
			// The larger the value the less packet loss you will see, but higher latency
			builder := samplebuilder.New(averageRtpPacketsPerFrame*5, pkt)
			for {
				select {

				case <-stop:
					return
				default:
					//读取RTP包
					rtp, err := remoteTrack.ReadRTP()
					if err != nil {
						if err == io.EOF {
							return
						}
						util.Errorf(err.Error())
					}

					//将RTP包放入builder对象里
					builder.Push(rtp)
					//迭代数据
					for s := builder.Pop(); s != nil; s = builder.Pop() {
						//向videoTrack里写入数据
						if err := (*videoTrack).WriteSample(*s); err != nil && err != io.ErrClosedPipe {
							util.Errorf(err.Error())
						}
					}
				}
			}
			//音频处理
		} else {
			*audioTrack, err = (*pc).NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), "audio", remoteTrack.Label())

			rtpBuf := make([]byte, 1400)
			for {
				select {
				case <-stop:
					return
				default:
					//读取音频数据
					i, err := remoteTrack.Read(rtpBuf)

					if err == nil {
						//将音频数据写入audioTrack
						(*audioTrack).Write(rtpBuf[:i])
						/* log.Print(rtpBuf[:i]) */
					} else {
						util.Infof(err.Error())
					}
				}
			}
		}
	})

	//设置远端SDP
	err = (*pc).SetRemoteDescription(offer)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	//创建应答Answer
	answer, err = (*pc).CreateAnswer(nil)
	//设置本地SDP
	err = (*pc).SetLocalDescription(answer)
	util.Infof("WebRTCEngine.CreateReceiver ok")
	return answer, err

}
