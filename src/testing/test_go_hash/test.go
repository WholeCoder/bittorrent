package main

import (
    "crypto/sha1"
    "fmt"
    "net/url"
)

func main() {
    a := "sha1 this string"

    h := sha1.New()

    h.Write([]byte(a))

    bs := h.Sum(nil)

    q := make(url.Values)
    q.Add("info", string(bs))

    fmt.Println(q.Encode())
}
