package main

import (
    "fmt"
    "time"
)

func main() {
    current := time.Now().Unix()
    fmt.Printf("\ntime int64 = %v\n", current)
    fmt.Printf("\ntime int   = %v\n", int(current))
}
