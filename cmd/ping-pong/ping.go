package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "time"
)

var (
    remoteAddress = flag.String("r", "127.0.0.1:3000", "remote address")
)

func sendPing() {
    ping, err := DarkYoung.NewClient(*remoteAddress)
    if err != nil {
        fmt.Println(err)
        return
    }

    for {
        request, err := ping.Request(1, []byte("ping"))
        if err != nil {
            fmt.Println(err)
            continue
        }

        go func() {
            data, status, err := request.Response(time.Second)
            if err != nil {
                fmt.Printf("data: %v %s, %v\n", status, data, err)
            }
        }()
    }
}


func main() {
    flag.Parse()
    sendPing()
}


