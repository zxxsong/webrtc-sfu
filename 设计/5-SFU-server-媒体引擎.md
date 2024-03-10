主要内容：编写webrtcengine.go

## 1. 媒体引擎的定义与配置
定义WebRTCEngine 结构体，包括：
- congfiguration
- webrtc.mediaEngine 媒体引擎说的就是这个，以及
- webrtc.api 媒体引擎相应的API。

此外还有一个变量 averageRtpPackagesPerFrame。

NewWebRTCEngine() 定义完后实例化对象，具体地：
1. 在cfg中设置SDP格式，ICEServers设置urls。
2. 设置媒体引擎mediaEngine的编码格式，视频和音频。
3. 配置实例化对象w的api。
4. return w

媒体引擎定义完成后，就在上一章节peer服务中加入定义的媒体引擎。

## 2. 媒体引擎接收数据
对于客户端发来的offer，应该怎样回复answer？A发给PC的过程

为此建立CreateRecevier（）函数，先接收客户端的数据，代码主要内容为：
- 先创立PC-A对象 *pc
- 同时添加Video和Audio，AddTranscevier（）
- （*pc）.OnTrack（），会回调远端的数据，即拿到A的数据，然后进行操作。**该步为关键操作** 具体逻辑操作看代码。 
- 之后设置远端SDP、创建Answer、本地SDP：SetRemoteDescription，CreateAnswer，SetLocalDescription。类似于1对1中的客户端和发布者answer一个信息。

## 3. 媒体引擎发送数据
在接收了由A发来的数据后，该怎样向B,C,D发送数据呢？PC发给B C D的过程

为此建立CreateSender（）函数，代码主要内容为：
- 先创建PC对象 *pc。
- 将VideoTrack和AudioTrack给到B C D。 （*pc）.addTrack
- 之后设置远端SDP、创建Answer、本地SDP：SetRemoteDescription，CreateAnswer，SetLocalDescription。

## 4. webrtcpeer.go中的SendPLI（）
对应上webrtcengine.go中的CreateRecevier（） go func部分。

