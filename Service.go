package DarkYoung

import (
    "context"
    proto "github.com/DGHeroin/DarkYoung/proto"
    "github.com/pkg/errors"
    "google.golang.org/grpc"
    "io"
    "net"
    "sync"
    "time"
)

var (
    errorTimeout = errors.New("Timeout")
)

type connectState int

const (
    connectStateInit connectState = iota
    connectStateConnecting
    connectStateConnected
    connectStateDisconnecting
    connectStateDisconnected
    connectStateError
)

type HandleFunc func(tag int32, request []byte) (response []byte, status int32)

type response struct {
    body   []byte
    status int32
    err    error
}
type Request struct {
    s     *service
    id    int32
    mutex sync.Mutex
    resp  chan response
}

func (req *Request) Response() ([]byte, int32, error) {
    resp := <-req.resp
    return resp.body, resp.status, resp.err
}

func (req *Request) Timeout(duration time.Duration) *Request {
    timeout := time.After(duration)
    go func() {
        select {
        case <-timeout:
            req.mutex.Lock()
            req.s.deleteRequest(req.id)
            req.resp <- response{err: errorTimeout}
            req.mutex.Unlock()
        }
    }()

    return req
}

type Service interface {
    Listen(address string) error
    Request(int32, []byte) (*Request, error)
}

type service struct {
    errChan       chan interface{}         // 错误chan
    state         connectState             // 连接状态
    server        *grpc.Server             // 作为服务端
    client        proto.Service_SendClient // 客户端实例
    conn          *grpc.ClientConn         // 作为客户端
    OnRequest     HandleFunc               // 请求处理函数
    requestId     int32                    // 请求的自增id
    requestMap    map[int32]*Request       // 请求映射表
    requestMutex  sync.RWMutex             // 请求锁定
    remoteAddress string                   // 远程地址
}

// 初始化service
func (s *service) init() {
    s.errChan = make(chan interface{}, 1)
    s.requestMap = make(map[int32]*Request)
    s.state = connectStateInit
}

// 创建服务器
func NewService(callback HandleFunc) Service {
    s := &service{}
    s.init()
    s.OnRequest = callback
    s.server = grpc.NewServer()
    proto.RegisterServiceServer(s.server, s)
    return s
}

// 创建客户端
func NewClient(address string) (Service, error) {
    s := &service{}
    s.init()
    s.remoteAddress = address
    if err := s.connectRemote(); err != nil {
        return nil, err
    }
    return s, nil
}

func (s *service) connectRemote() error {

    s.state = connectStateConnecting
    conn, err := grpc.Dial(s.remoteAddress, grpc.WithInsecure())
    if err != nil {
        s.state = connectStateError
        return err
    }

    s.conn = conn
    client := proto.NewServiceClient(conn)
    // 创建双向数据流
    stream, err := client.Send(context.Background())
    if err != nil {
        s.state = connectStateError
        return err
    }
    s.client = stream
    s.state = connectStateConnected
    // 运行接收
    go s.clientRecv()
    return nil
}

// 客户端接收信息
func (s *service) clientRecv() {
    for {
        msg, err := s.client.Recv()
        if err == io.EOF {
            s.state = connectStateDisconnected
            break // 收到服务端的结束信号
        }
        if err != nil {
            s.state = connectStateError
            go func() { s.errChan <- err }()
            return // 错误
        }
        s.requestMutex.RLock()
        req, ok := s.requestMap[msg.Id]
        s.requestMutex.RUnlock()
        if ok {
            go func() {
                defer func() {
                    if e := recover(); e != nil {
                        s.state = connectStateError
                        go func() { s.errChan <- err }()
                    }
                }()
                req.mutex.Lock()
                s.deleteRequest(msg.Id)
                req.resp <- response{body: msg.Body, status: msg.Status}
                req.mutex.Unlock()
            }()

        }
    }
}

func (s *service) Send(stream proto.Service_SendServer) error {
    ctx := stream.Context()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // client stop sig
        default:
            msg, err := stream.Recv()
            if err == io.EOF {
                s.state = connectStateDisconnected
                return nil // close by client
            }
            if err != nil {
                s.state = connectStateError
                go func() { s.errChan <- err }()
                return err // read data error, network error
            }
            responseData, status := s.OnRequest(msg.Tag, msg.Body)
            responseMsg := &proto.Response{Id: msg.Id, Body: responseData, Status: status}
            if err := stream.Send(responseMsg); err != nil {
                s.state = connectStateError
                go func() { s.errChan <- err }()
                return err // send error
            }

        }
    }
}

func (s *service) Listen(listenAddress string) error {
    address, err := net.Listen("tcp", listenAddress)
    if err != nil {
        return err
    }
    s.state = connectStateConnected
    return s.server.Serve(address)
}

func (s *service) Request(tag int32, data []byte) (*Request, error) {
    if s.state != connectStateConnected {
        if err := s.connectRemote(); err != nil {
            return nil, err
        }
    }

    s.requestMutex.Lock()

    Id := s.requestId
    s.requestId++
    result := &Request{s: s, id: Id, resp: make(chan response)}
    s.requestMap[Id] = result

    s.requestMutex.Unlock()

    request := &proto.Request{Id: Id, Tag: tag, Body: data}
    err := s.client.Send(request)
    if err != nil { // 传输错误, 删除请求记录
        s.state = connectStateError
        s.deleteRequest(Id)
    }
    return result, err
}

func (s *service) deleteRequest(id int32) {
    s.requestMutex.Lock()
    delete(s.requestMap, id)
    s.requestMutex.Unlock()
}
