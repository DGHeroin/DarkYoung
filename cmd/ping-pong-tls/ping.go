package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/DarkYoung"
)

var (
    remoteAddress = flag.String("r", "127.0.0.1:3000", "remote address")
    connectionNum = flag.Int("c", 1, "connection num")
)

func sendPing() {
    ping, err := DarkYoung.NewClient(*remoteAddress,
        DarkYoung.WithClientTLS("client/ca.pem", "client/cert.pem", "client/key.pem"))
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
        //go func() {
            data, status, err := request.Response()
            if err != nil {
                fmt.Printf("data: %v %s, %v\n", status, data, err)
            }
        //}()
    }
}

func main() {
    flag.Parse()
    for i := 0; i < *connectionNum; i++ {
        go sendPing()
    }
    select{}
}


