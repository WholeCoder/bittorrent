package main

import (
    "fmt"
    "math"
)

const REQUEST_SIZE float64 = 16384.0 // 2^14

func main() {
    fmt.Printf("\npow val:  %v\n\n", REQUEST_SIZE)

    std_piece_blocks := int(math.Ceil(float64(2.5 * 16384.0 / REQUEST_SIZE)))
    fmt.Printf("\nstd val: %v\n\n", std_piece_blocks)
}
