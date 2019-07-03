package DarkYoung

import (
    "context"
    "fmt"
    proto "github.com/DGHeroin/DarkYoung/proto"
    "google.golang.org/grpc"
    "io"
    "net"
    "sync"
    "sync/atomic"
)

type server struct {
    ctx           context.Context
    listener      net.Listener      // rpc listener
    listenAddress string            // 监听地址
    server        *grpc.Server      // 作为服务端
    option        serverOption      // 配置
    clientId      int32             // 没连接一个新连接自增 1
    clients       map[int32]*client // 客户端
    clientsMutex  sync.RWMutex
}

// 关闭服务器
func (s *server) Close() (err error) {
    if s.listener != nil {
        err = s.listener.Close()
        s.listener = nil
    }
    return nil
}

// 初始化service
func (s *server) init(address string) error {
    var err error
    if s.option.withTLS == false {
        s.server = grpc.NewServer()
    } else { // 使用 TLS X509
        creds, err := loadCredentials(s.option.caPath, s.option.certPath, s.option.keyPath)
        if err != nil {
            return err
        }
        s.server = grpc.NewServer(grpc.Creds(creds))
    }
    s.listenAddress = address
    proto.RegisterServiceServer(s.server, s)
    // 监听
    s.listener, err = net.Listen("tcp", address)
    if err != nil {
        return err
    }
    go func(ctx context.Context) {
        if err := s.server.Serve(s.listener); err != nil {
            fmt.Println("server Serve error", err)
        }
    }(s.ctx)
    return nil
}

// 服务端接收到消息
func (s *server) Send(stream proto.Service_SendServer) error {
    // 新的客户端连接
    cli := newAcceptClient(s.ctx)
    ctx := stream.Context()
    cli.id = atomic.AddInt32(&s.clientId, 1)
    s.clientsMutex.Lock()
    s.clients[cli.id] = cli
    s.clientsMutex.Unlock()

    cli.stream = stream
    cli.changeState(connectionStateConnected)

    defer func() {
        cli.streamM.Lock()
        cli.stream = nil
        cli.streamM.Unlock()

        s.clientsMutex.Lock()
        delete(s.clients, cli.id)
        s.clientsMutex.Unlock()
    }()

    // OnClientConnected
    if s.option.onConnected != nil {
        s.option.onConnected(cli.id)
    }
    defer func() {
        // OnClientClosed
        if s.option.onClosed != nil {
            s.option.onClosed(cli.id)
        }
    }()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // client stop sig
        default:
            msg, err := cli.RecvStream()
            if err == io.EOF {
                cli.changeState(connectionStateDisconnected)
                return nil // close by client
            }
            if err != nil {
                cli.changeState(connectionStateError)
                return err // read data error, network error
            }
            // OnClientData
            if s.option.onMessage != nil {
                responseData, status := s.option.onMessage(cli.id, msg.Tag, msg.Body)
                responseMsg := &proto.Response{Id: msg.Id, Body: responseData, Status: status, Type:0}

                if err := cli.SendResponse(responseMsg); err != nil {
                    cli.changeState(connectionStateError)
                    return err
                }
            }
        }
    }
}

func (s *server) String() string {
    s.clientsMutex.RLock()
    n := len(s.clients)
    s.clientsMutex.RUnlock()

    return fmt.Sprintf("{\"address\":\"%v\", \"clients\":%v}", s.listenAddress, n)
}

func (s *server) CloseClient(id int32) (err error) {
    s.clientsMutex.Lock()
    if cli, ok := s.clients[id]; ok {
        delete(s.clients, id)
        err = cli.Close()
    } else {
        err = errorClientNotFound
    }
    s.clientsMutex.Unlock()
    return
}

func (s *server) SendTo(id int32, tag int32, body []byte) (err error) {
    s.clientsMutex.RLock()
    cli, ok := s.clients[id]
    s.clientsMutex.RUnlock()
    if !ok || cli == nil {
        err = errorClientNotFound
        return
    }
    return cli.PushMessage(tag, body)
}