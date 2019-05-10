package DarkYoung

import (
    "context"
    proto "github.com/DGHeroin/DarkYoung/proto"
    "sync"
    "time"
)

type response struct {
    body   []byte
    status int32
    err    error
}

type Request struct {
    ctx   context.Context
    s     interface{}
    id    int32
    mutex sync.Mutex
    resp  chan response
}

func (req *Request) Response(durations ...time.Duration) ([]byte, int32, error) {
    duration := time.Second * 120
    if len(durations) > 0 {
        duration = durations[0]
    }

    defer func() {
        req.mutex.Lock()
        req.resp = nil
        req.mutex.Unlock()
    }()
    select {
    case <-req.ctx.Done():
        return nil, 0, errorTimeout
    case resp := <-req.resp:
        req.mutex.Lock()
        if req.resp == nil {
            return nil, 0, errorTimeout
        }
        req.mutex.Unlock()

        return resp.body, resp.status, resp.err
    case <-time.After(duration):
        return nil, 0, errorTimeout
    }
}

type requestManager struct {
    ctx          context.Context    //
    requestId    int32              // 请求的自增id
    requestMap   map[int32]*Request // 请求映射表
    requestMutex sync.RWMutex       // 请求锁定
}

func (r *requestManager) init(ctx context.Context) {
    r.requestMap = make(map[int32]*Request)
    r.ctx = ctx
}

func (r *requestManager) Get(id int32) *Request {
    r.requestMutex.Lock()
    if req, ok := r.requestMap[id]; ok {
        r.requestMutex.Unlock()
        return req
    }
    r.requestMutex.Unlock()
    return nil
}

func (r *requestManager) Clear() {
    r.requestMutex.Lock()
    r.requestMap = make(map[int32]*Request)
    r.requestMutex.Unlock()
}

func (req *requestManager) deleteRequest(id int32) {
    req.requestMutex.Lock()
    if r, ok := req.requestMap[id]; ok {
        r.ctx.Done()
    }
    delete(req.requestMap, id)

    req.requestMutex.Unlock()
}

func (req *requestManager) newPBRequest(id, tag int32, data []byte) *proto.Request {
    return &proto.Request{Id: id, Tag: tag, Body: data}
}

func (req *requestManager) Request() *Request {
    req.requestMutex.Lock()
    Id := req.requestId
    req.requestId++
    result := &Request{s: req, id: Id, resp: make(chan response), ctx:req.ctx}
    req.requestMap[Id] = result
    req.requestMutex.Unlock()
    return result
}
