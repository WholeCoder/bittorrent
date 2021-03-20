package main

import (
    "fmt"
    "log"
    "bytes"
    "github.com/lunixbochs/struc"
)

type Example struct {
    Length byte `struc:"big"`
    Protocol string `struc:"[19]int8"`
    Space [8]byte
    Info_hash string `struc:"[20]int8"`
    Peer_id string `struc:"[20]int8"`

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
    o := &Example{}
    err = struc.Unpack(&buf, o)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Unpacked struc:  %#v\n\n", o)
    fmt.Printf("\nProtocol %v\n", o.Protocol)
}
