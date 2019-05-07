package main

import (
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "log"
    "sync/atomic"
    "time"
)

type Pong struct {}

var (
    qps = int64(0)
)

func (p *Pong) OnRequest(tag int32, request []byte) (response []byte, status int32) {
    response = []byte("pongongongongong")
    atomic.AddInt64(&qps, 1)
    //if tag == 16 {
    //    r := rand.Intn(200)
    //    time.Sleep(time.Millisecond* time.Duration(r))
    //    fmt.Println("random:", r)
    //    status = int32(r)
    //}
    return
}

func main()  {
    go func() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            if qps != 0 {
                log.Printf("qps:%v\n", qps)
            }
            atomic.StoreInt64(&qps, 0)
        }
    }()

    pong := &Pong{}
    if err := DarkYoung.NewService(pong.OnRequest).Listen(":3000"); err != nil {
        fmt.Println(err)
    }
}

