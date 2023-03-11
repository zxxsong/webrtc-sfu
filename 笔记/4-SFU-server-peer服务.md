主要内容：编写webrtcpeer.go；讲**转发**的基础:webrtcpeer.go、webrtcengine.go

## 1.转发原理

通信过程如下所示：

**A --> SFU --> B**

sfu架构中，通信过程，服务器端是作为应答方的，所以应答方使用Answer。具体可看 2图

Answer 分为 AnswerSender和AnswerRecevier。

可以看出 A 是AnswerSender，B 是 AnswerRecevier。在没创建A和B这两个用户的时候，就要创建webrtcpeer对象。

## 2.编写webrtcpeer.go peer定义

### 2.1 定义WebrtcPeer结构体
包括：
- ID，string
- PC，*webrtc.PeerConnection (pion库的api) 实例化PC相当于连接过程中的PC-A,PC-B,PC-C，同时还有video/audio 轨道。
- VideoTrack和AudioTrack，每个PC都有一个VideoTrack和AudioTrack, *webrtc.Track (pion库的api)。
- stop chan int 通道，停止使用连接的，整型。
- pli chan int 关键帧丢包重传。

定义完后实例化一个对象。

## 3.编写wertcpeer.go 响应客户端

### 3.1 服务器端要分别响应sender和recevier
- 响应发送方 AnswerSender()
- 响应接收方 AnswerRecevier()

### 3.2 关键帧丢包重传
- SendPLI() 通过通道的方式，让引擎开始接收数据。



