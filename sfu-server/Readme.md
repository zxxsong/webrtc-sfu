https连接 升级成 wss连接
wss
openssl : 如果想自己生成，就用openssl

https://www.xxx.com


A-->SFU-->B

AnswerSender A
AnswerReceiver B

加入房间
离开房间
发布
订阅

要先建立websocket连接，即先访问下面链接
https://0.0.0.0:3000/ws
https://localhost:3000/ws

别人访问，更换为服务器的ip地址

项目有两个服务，一个是wss连接的信令及sfu服务，一个是node.js的应用程序服务