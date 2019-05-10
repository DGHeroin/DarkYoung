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

func (p *Pong) OnRequest(id int64, tag int32, request []byte) (response []byte, status int32) {
    response = []byte("p")
    atomic.AddInt64(&qps, 1)
    return
}

func (p *Pong) OnNew(id int64) {
    fmt.Println("接收到新连接", id)
}

func (p *Pong) OnClosed(id int64) {
    fmt.Println("连接断开", id)
}

func startPong() {
    pong := &Pong{}
    if server, err := DarkYoung.NewServer(*serveAddress,
        DarkYoung.WithServerOnAccept(pong.OnNew),
        DarkYoung.WithServerOnMessage(pong.OnRequest),
        DarkYoung.WithServerOnClosed(pong.OnClosed)); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("服务已经启动", server)
        select {}
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
    data, status, err := request.Response(time.Second)
    if err != nil {
        fmt.Printf("data: %v %s, %v\n", status, data, err)
    }
}
```

### TLS使用方式
服务端
```
// 创建server时, 只需传入
DarkYoung.WithServerTLS("server/ca.pem", "server/server-cert.pem", "server/server-key.pem")
```

客户端端
```
// 创建client时, 只需传入
DarkYoung.WithClientTLS("server/ca.pem", "server/server-cert.pem", "server/server-key.pem")
```