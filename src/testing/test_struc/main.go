package main

import (
    "fmt"
    "log"
    "bytes"
    "github.com/lunixbochs/struc"
)

type Example struct {
    Length byte `struc:"big"`
    Protocol string
    Space [8]byte
    Info_hash string
    Peer_id string
}

func main() {
    var buf bytes.Buffer
    t := &Example{byte(19), "BitTorrent protocol", [8]byte{0,0,0,0,0,0,0,0},"01234567890123456789", "01234567890123456789"}
    err := struc.Pack(&buf, t)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\n%#v\n\n", t)
    fmt.Printf("%#v\n\n", buf.Bytes())
    //o := &Example{}
    //err = struc.Unpack(&buf, o)
}
