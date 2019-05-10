package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "log"
    "sync/atomic"
    "time"
)

type Pong struct {}

var (
    qps = int64(0)
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

func main()  {
    flag.Parse()
    go func() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            if qps != 0 {
                log.Printf("qps:%v(%v)\n", size(qps), qps)
            }
            atomic.StoreInt64(&qps, 0)
        }
    }()

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

