package DarkYoung

// 服务器配置
type serverOption struct {
    onMessage func(id int64,
        tag int32,
        request []byte) (response []byte, status int32) // 请求消息
    onConnected func(id int64)                          // 连接上
    onClosed    func(id int64)                          // 连接关闭
    withTLS     bool                                    // 使用TLS
    certPath    string                                  // TLS cert 路径
    keyPath     string                                  // TLS key 路径
    caPath      string                                  // TLS ca 路径
}

type ServerOptionFunc func(*serverOption)

func defaultServerOption() serverOption {
    return serverOption{}
}

// Option 连接上
func WithServerOnAccept(onConnected func(id int64)) ServerOptionFunc {
    return func(option *serverOption) {
        option.onConnected = onConnected
    }
}

//
func WithServerOnMessage(onMessage func(id int64, tag int32, request []byte) (response []byte, status int32)) ServerOptionFunc {
    return func(option *serverOption) {
        option.onMessage = onMessage
    }
}

// Option 连接上
func WithServerOnClosed(onClosed func(id int64)) ServerOptionFunc {
    return func(option *serverOption) {
        option.onClosed = onClosed
    }
}

// server 使用 TLS
func WithServerTLS(ca string, cert string, key string) ServerOptionFunc {
    return func(option *serverOption) {
        option.withTLS = true
        option.caPath = ca
        option.certPath = cert
        option.keyPath = key
    }
}

//
// Client
//
type clientOption struct {
    withTLS  bool
    certPath string
    keyPath  string
    caPath   string
}
type ClientOptionFunc func(*clientOption)
func defaultClientOption() clientOption {
    return clientOption{}
}
// client使用 TLS
func WithClientTLS(ca string, cert string, key string) ClientOptionFunc {
    return func(option *clientOption) {
        option.withTLS = true
        option.certPath = cert
        option.keyPath = key
        option.caPath = ca
    }
}
