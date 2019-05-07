package main

import (
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "time"
)

func sendPing() {
    ping, err := DarkYoung.NewClient("127.0.0.1:3000")
    if err != nil {
        fmt.Println(err)
        return
    }

    for range time.Tick(time.Second){
        request, err := ping.Request(1, []byte("ping"))
        if err != nil {
            fmt.Println(err)
            continue
        }
        go func() {
            data, status, err := request.Response()
            if err != nil {
                fmt.Printf("data: %v %s, %v\n", status, data, err)
            }
        }()
    }
}

var (
    qps = false
)

func main()  {
    if qps {
       //testQPSPing()
    } else {
        sendPing()
    }

}


