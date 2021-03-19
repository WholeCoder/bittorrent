package main

import "fmt"

type PeerMessage interface {
    Multiply(a float32, b float32) float32
}

type Arith struct {
}

func (Arith) Multiply(a float32, b float32) float32 {
    return a * b
}

func main() {
    pMessage := Arith
    result := (pMessage).Multiply(Arith{}, 15, 25)
    fmt.Println(result)
}
