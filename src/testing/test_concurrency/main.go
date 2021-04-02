package main

import (
	"fmt"
    "time"
)

type Test struct {
    data string
}

func (t *Test) print() {
	fmt.Println(t.data)
}

func main() {
    t := Test{data: "Ruben is Awesome!"}
	go t.print()
    duration := time.Duration(10)*time.Second
    time.Sleep(duration)
}

