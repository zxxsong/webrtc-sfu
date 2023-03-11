主要内容：编写room.go，主要是房间中建立连接时的peer方面。

## 1. 添加peer： addWebRTCPeer():

1. 先判断是不是sender，即是不是发布者。
2. 如果是sender，则对pubPeer加锁，之后defer pubPeerLock.Unlock解锁。
3. 判断pubPeers中传入id是否为空，不为空先停止，再用media.NewWebRTCPeer(id)创建。
4. 不是sender，即不是发布者，是订阅者，执行相同逻辑，将变量和方法换为subPeer那一套。

## 2. 获取Peer：getWebRTCPeer():

- 与add相同的逻辑，最后换成return对应id的Peer。

从中可以看出，pubPeer对应sender，subPeer对应receiver。

## 3. 删除Peer：delWebRTCPeer():

1. 判断是不是sender。
2. 是sender，pubPeer加锁，defer延迟解锁。
3. 判断要删除的peer id是否不为空，对应id下的PC是否不为空，均不为空，则pubPeer.PC.Close()，pubPeer.Stop()。
4. 最后所有都停止后delete()。
5. 不是sender，逻辑相同，改成subPeer那一套。

## 4. 关键帧丢包重传
对于room.go，还需要写sendPLI（）。

## 5. 发送消息：sendMessage()
当一个人登录进房间后，通知其他人登录进来了。sendMessage()是room类的必须方法。
sendMessage()中迭代循环room中的所有user，除去自己，发出通知。 

## 6. 房间关闭：Close()
1. 加锁，derfer解锁
2. 迭代所有pubpeer，将其stop。

## 7. 应答客户端的处理：answer()
1. 如果是sender，先将offer响应，同时记为answer，使用answerSender。
2. 不是sender，则先找到发布者的pubPeer，将video和audio转给订阅者，使用AnswerRecevier，代码中的sleep是等待video和audio。

流程：对于不是sender的情况，获取到pubPeer的videoTrack和audioTrack，通过AnswerRecevier给到引擎，引擎再给到PC-B PC-C PC-D。


