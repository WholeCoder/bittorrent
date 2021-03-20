package main

import (
    "bytes"
    "encoding/gob"
    "fmt"
    "log"
)

type S struct {
    Length byte
    Protocol string
    Pad [8]byte
    Info_hash string
    Peer_id string
}

type P struct {
    X, Y, Z int
    Name    string
}

type Q struct {
    X, Y *int32
    Name string
}

func main() {
    // Initialize the encoder and decoder.  Normally enc and dec would be
    // bound to network connections and the encoder and decoder would
    // run in different processes.
    var network bytes.Buffer        // Stand-in for a network connection
    enc := gob.NewEncoder(&network) // Will write to network.
    dec := gob.NewDecoder(&network) // Will read from network.
    // Encode (send) the value.
    err := enc.Encode(S{byte(19), "BitTorrent protocol", [8]byte{0,0,0,0,0,0,0,0}, "01234567890123456789","01234567890123456789"})
    // err := enc.Encode(P{3, 4, 5, "Pythagoras"})
    if err != nil {
        log.Fatal("encode error:", err)
    }

    // HERE ARE YOUR BYTES!!!!
    fmt.Println("first byte =", network.Bytes()[0])

    // Decode (receive) the value.
    var s S
    err = dec.Decode(&s)
    if err != nil {
        log.Fatal("decode error:", err)
    }
    fmt.Printf("%v, %v, %v\n", s.Protocol, s.Info_hash, byte(19))
}
