package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/DarkYoung"
    "log"
    "time"
)

var (
    remote = flag.String("r", "127.0.0.1:3002", "remote address")
    //stringData = flag.Bool("str", true, "send string")
    tag = flag.Int("tag", 10, "message tag")
    stringData = flag.String("data", "helloworld", "string content")
)

func main()  {
    flag.Parse()
    cli, err := DarkYoung.NewClient(*remote)
    if err != nil {
        fmt.Println(err)
        return
    }
    if req, err := cli.Request(int32(*tag), []byte(*stringData)); err != nil {
        fmt.Println(err)
        return
    } else {
        data, status, err := req.Response(time.Second * 10)
        if err != nil {
            fmt.Println(err)
        } else {
            log.Printf("response:%d %s", status, data)
        }
    }
}
