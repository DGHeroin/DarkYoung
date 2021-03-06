package DarkYoung

import (
    "context"
    "fmt"
    proto "github.com/DGHeroin/DarkYoung/proto"
    "google.golang.org/grpc"
    "io"
    "sync"
)

type connectionType int

const (
    connectionTypeInitiative connectionType = iota // 主动发起的连接
    connectionTypePassive                          // 服务端接收到的客户端
)

type client struct {
    ctx           context.Context
    id            int32                    // 运行时id
    connState     connectionState          // 连接状态
    client        proto.Service_SendClient // 客户端实例
    conn          *grpc.ClientConn         // 作为客户端
    remoteAddress string                   // 远程地址
    mgr           requestManager           // 请求管理
    connectType   connectionType           // 连接
    option        clientOption             // 客户端配置
    onPush        func(int32, []byte)      // 消息推送回调
    stream        proto.Service_SendServer
    streamM       sync.RWMutex
}

func (cli *client) init() {
    cli.mgr.init(cli.ctx)
    cli.changeState(connectionStateInit)
}

func (cli *client) Request(tag int32, data []byte) (req *Request, err error) {
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        if cli.connState != connectionStateConnected {
            if err = cli.connectRemote(); err != nil {
                return
            }
        }
        if cli.client == nil {
            err = errorNotConnected
            return
        }

        req = cli.mgr.Request()
        err = cli.client.Send(cli.mgr.newPBRequest(req.id, tag, data))
        if err != nil { // 传输错误, 删除请求记录
            cli.changeState(connectionStateError)
            cli.mgr.deleteRequest(req.id)
        }
    }()
    wg.Wait()
    return req, err
}

func (c *client) Close() (err error) {
    c.changeState(connectionStateDisconnecting)
    if c.client != nil {
        err = c.client.CloseSend()
        c.client = nil
    }
    c.mgr.Clear()
    c.changeState(connectionStateDisconnected)
    return nil
}

// 客户端接收信息
func (cli *client) clientRecv() {
    defer func() {
        if err := cli.Close(); err != nil {
            fmt.Println(err)
        }
    }()
    for {
        msg, err := cli.client.Recv()
        if err == io.EOF {
            cli.changeState(connectionStateDisconnected)
            break // 收到服务端的结束信号
        }
        if err != nil {
            cli.changeState(connectionStateError)
            return // 错误
        }
        switch msg.Type {
        case 0:
            // request
            req := cli.mgr.Get(msg.Id)
            if req != nil {
                go func() {
                    defer func() {
                        if e := recover(); e != nil {
                            cli.changeState(connectionStateError)
                        }
                    }()
                    req.mutex.Lock()
                    cli.mgr.deleteRequest(req.id)
                    req.resp <- response{body: msg.Body, status: msg.Status}
                    req.mutex.Unlock()
                }()
            }
        case 1:
            // push
            if cli.onPush != nil {
                cli.onPush(msg.Tag, msg.Body)
            }
        default:
            // not support
        }

    }
}

// 客户端改变状态
func (cli *client) changeState(state connectionState) {
    cli.connState = state
}

// 连接到远程
func (cli *client) connectRemote() error {
    cli.changeState(connectionStateConnecting)
    var (
        conn *grpc.ClientConn
        err  error
    )
    if cli.option.withTLS == false {
        conn, err = grpc.Dial(cli.remoteAddress, grpc.WithInsecure())
    } else { // 使用 TLS
        creds, err := loadClientCredentials(cli.option.caPath, cli.option.certPath, cli.option.keyPath, "127.0.0.1")
        if err != nil {
            return err
        }
        conn, err = grpc.Dial(cli.remoteAddress, grpc.WithTransportCredentials(creds))
    }
    if err != nil {
        cli.changeState(connectionStateError)
        return err
    }
    cli.conn = conn
    client := proto.NewServiceClient(conn)
    // 创建双向数据流
    stream, err := client.Send(context.Background())
    if err != nil {
        cli.changeState(connectionStateError)
        return err
    }

    cli.client = stream
    cli.changeState(connectionStateConnected)
    // 运行接收
    go cli.clientRecv()
    return nil
}

func (cli *client) HandlePushMessageFunc(cb func(int32, []byte)) {
    cli.onPush = cb
}

func (cli *client) PushMessage(tag int32, body []byte) (err error) {
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        cli.streamM.RLock()
        if cli.stream == nil {
            err = errorNotConnected
            return
        }
        err =  cli.stream.Send(&proto.Response{Type: 1, Tag: tag, Body: body})
        cli.streamM.RUnlock()

        if err != nil {
            cli.streamM.Lock()
            cli.stream = nil
            cli.streamM.Unlock()
        }
    }()
    wg.Wait()
    return
}

func (cli *client) RecvStream() (*proto.Request, error) {
    cli.streamM.RLock()
    msg, err := cli.stream.Recv()
    cli.streamM.RUnlock()
    return msg, err
}

func (cli *client) SendResponse(rsp *proto.Response) (err error) {
    cli.streamM.RLock()
    err = cli.stream.Send(rsp)
    cli.streamM.RUnlock()
    return
}