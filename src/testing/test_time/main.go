package main

import (
    "fmt"
    "time"
)

func main() {
    current := time.Now().Unix()
    fmt.Printf("\ntime = %v\n", current)
}
