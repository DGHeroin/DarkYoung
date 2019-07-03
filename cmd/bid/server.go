package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "log"
    "sync"
    "sync/atomic"
    "time"
)

type clientState struct {
    id int32
}

type Pong struct {
    clientM sync.RWMutex
    clients map[int32]*clientState
    server DarkYoung.Server
}

var (
    qps          = int64(0)
    pushQPS  = int64(0)
    serveAddress = flag.String("l", ":3000", "listen address")
)

func size(n int64) string {
    unit := ""
    if n > 1000 {
        n = n / 1000
        unit = "K"
    }
    if n > 1000 {
        n = n / 1000
        unit = "M"
    }
    if n > 1000 {
        n = n / 1000
        unit = "G"
    }

    return fmt.Sprintf("%v%v", n, unit)
}

func (p *Pong) init() {
    p.clients = make(map[int32]*clientState)
}

func (p *Pong) OnMessage(id int32, tag int32, request []byte) (response []byte, status int32) {
    response = []byte("p")
    atomic.AddInt64(&qps, 1)
    return
}

func (p *Pong) OnConnected(id int32) {
    fmt.Println("接收到新连接", id)
    p.clientM.Lock()
    p.clients[id] = &clientState{id:id}
    p.clientM.Unlock()

    go func(id int32) {
        count := 0
        for {
            if err := p.server.SendTo(id, 100, []byte("sever ttl")); err != nil {
                fmt.Println("server send ttl error:", err)
                break
            }
            count++
            atomic.AddInt64(&pushQPS, 1)
            time.Sleep(time.Nanosecond)
        }
        log.Println("结束推送")

    }(id)
}

func (p *Pong) OnClosed(id int32) {
    fmt.Println("连接断开", id)
    p.clientM.Lock()
    delete(p.clients, id)
    p.clientM.Unlock()
}

func main() {
    flag.Parse()
    go func() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            if qps != 0 {
                fmt.Printf("request qps:%v(%v) push qps: %v(%v)\n", size(qps), qps, size(pushQPS), pushQPS)
            }
            atomic.StoreInt64(&qps, 0)
            atomic.StoreInt64(&pushQPS, 0)
        }
    }()

    pong := &Pong{}
    pong.init()
    if server, err := DarkYoung.NewServer(*serveAddress, DarkYoung.WithServerHandler(pong)); err != nil {
        fmt.Println(err)
    } else {
        fmt.Println("服务已经启动", server)
        pong.server = server
        select {}
    }
}
