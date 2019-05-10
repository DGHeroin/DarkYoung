package DarkYoung

import (
    "context"
    "github.com/pkg/errors"
)

// 错误
var (
    errorTimeout = errors.New("Timeout")
    errorNotConnected = errors.New("not connected")
)

// 连接状态
type connectionState int
const (
    connectionStateInit          connectionState = iota // 初始化
    connectionStateConnecting                           // 连接中
    connectionStateConnected                            // 已连接
    connectionStateDisconnecting                        // 断开中
    connectionStateDisconnected                         // 已断开
    connectionStateError                                // 错误
)

type Client interface {
    Request(int32, []byte) (*Request, error)
    Close() error
}

type Server interface {
    Close() error
}

// 创建服务器
func NewServer(address string, opts ...ServerOptionFunc) (Server, error) {
    o := defaultServerOption()
    for _, opt := range opts {
        opt(&o)
    }
    var err error
    s := &server{}
    s.ctx = context.Background()
    s.option = o
    err = s.init(address)
    return s, err
}

// 创建客户端
func NewClient(address string, opts ...ClientOptionFunc) (Client, error) {
    o := defaultClientOption()
    for _, opt := range opts {
        opt(&o)
    }

    cli := &client{}
    cli.ctx = context.Background()
    cli.init()
    cli.connectType = connectionTypeInitiative
    cli.remoteAddress = address
    cli.option = o

    if err := cli.connectRemote(); err != nil {
        return nil, err
    }
    return cli, nil
}

// 服务器接受连接
func newAcceptClient(ctx context.Context) *client {
    cli := &client{}
    cli.ctx = ctx
    cli.init()
    cli.connectType = connectionTypePassive
    return cli
}

