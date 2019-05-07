# DarkYoung
一个golang grpc双向流消息的简单封装.

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
