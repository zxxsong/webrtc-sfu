## 1.sokect流程

具体可回顾一对一案例。
用conn基础库，作消息的收和发，收就是客户端发给服务器端，发就是服务器端发给客户端。

**conn.go** 代码内容：

- 定义了websokect连接的结构体，同时实例化一个websokect连接。
- 对于接收数据，用ReadMessage（）向外派发，将数据放入通道里，循环接收通道里的数据，将消息派发出去。
- 对于发送数据同理，使用send（）。
- 最后关闭websokect连接，使用close（）关闭websokect连接。

**server.go** 代码内容：

- 定义了SFUServerConfig结构体，作为服务器配置，结构体包括IP、端口和证书文件，同时可以初始化一个SFUServer的配置。
- 定义了SFUServer结构体，作为服务器，其中将https升级为wss。
- 最后通过bind（）启动https服务，由于升级，启动wss服务。bind（）由main.go调用。

**main.go** 代码内容：
- 实例化各种对象，并作配置--config，最后使用wssserver绑定config，调用conn.go和server.go，启动服务。


