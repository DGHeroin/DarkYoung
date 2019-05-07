# DarkYoung
一个golang grpc双向流消息的简单封装.

## 消息定义
```go
Tag   请求的消息tag
Body  请求的消息体
```

Tag可以很好的区别消息的类别.</br>
例如:</br>
tag 1 表示用户注册消息</br>
tag 2 表示用户登录消息</br>
</br>
对应的tag就可以采用对应的消息格式进行 Unmarshal</br>



## 启动服务器
```go
type Pong struct {}

func (p *Pong) OnRequest(tag int32, request []byte) (response []byte, status int32) {
    response = []byte("pongongongongong")
}

func startPong() {
    pong := &Pong{}
    if err := DarkYoung.NewService(pong.OnRequest).Listen(":3000"); err != nil {
        fmt.Println(err)
    }
}
```

### 客户端发消息
```go
func startPing() {
    // 连接服务器
    ping, err := DarkYoung.NewClient("127.0.0.1:3000")
    // 请发送请求
    request, err := ping.Request(1, []byte("ping"))
    data, status, err := request.Response()
    if err != nil {
        fmt.Printf("data: %v %s, %v\n", status, data, err)
    }
}
```
