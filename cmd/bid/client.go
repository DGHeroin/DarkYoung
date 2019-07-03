package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "sync/atomic"
    "time"
)

var (
    remoteAddress = flag.String("r", "127.0.0.1:3000", "remote address")
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

func sendPing() {
    ping, err := DarkYoung.NewClient(*remoteAddress)
    if err != nil {
        fmt.Println(err)
        return
    }
    pushQPS := int64(0)
    recvQPS := int64(0)
    ping.HandlePushMessageFunc(func(tag int32, body []byte) {
        atomic.AddInt64(&pushQPS, 1)
    })

    // print qps
    go func() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            fmt.Printf("client recv qps:%v(%v) send qps:%v(%v)\n", size(pushQPS), pushQPS, size(recvQPS), recvQPS)
            atomic.StoreInt64(&pushQPS, 0)
            atomic.StoreInt64(&recvQPS, 0)
        }
    }()
    for {
        request, err := ping.Request(1, []byte("ping"))
        if err != nil {
            fmt.Println(err)
            continue
        }

        data, status, err := request.Response(time.Second)
        if err != nil {
            fmt.Printf("data: %v %s, %v\n", status, data, err)
        }
        atomic.AddInt64(&recvQPS, 1)
    }
}

func main() {
    flag.Parse()
    sendPing()
}


