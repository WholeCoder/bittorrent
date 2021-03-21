package main

import (
    "fmt"
    "encoding/binary"
)

func main() {
    tryInterface(10)
    res := []interface{}{}
    fmt.Println(res)
    b := []byte("test")
    fmt.Println(b)
    str2 := string(b[:])
    fmt.Printf("String: %v %T\n",str2[0], str2[0])
    fmt.Println("string converted back to []byte: ",[]byte(str2))
    var mySlice = []byte{0, 0, 0, 0, 0, 0, 1, 255}
    data := int(binary.BigEndian.Uint64(mySlice))
    fmt.Println(data)

    for index, value := range str2 {
        fmt.Println("str2[", index, "] = ", value)
    }
    m := make(map[interface{}]interface{})
    m["Ruben"] = 10
    m["Ruth"] = 12
    printType(m)

    l := []interface{}{}
    l = append(l, "test string")
    l = append(l, 42)
    printType(l)

    h := (uint32)(255)
    a := make([]byte, 4)
    binary.BigEndian.PutUint32(a, h)
    fmt.Printf("\nbyte array is:  %#v\n", a)
}

func printType(myInterface interface{}) {
    switch v := myInterface.(type) {
    case int:
        // v is an int here, so e.g. v + 1 is possible.
        fmt.Printf("Integer: %v", v)
    case float64:
        // v is a float64 here, so e.g. v + 1.0 is possible.
        fmt.Printf("Float64: %v", v)
    case string:
        // v is a string here, so e.g. v + " Yeah!" is possible.
        fmt.Printf("String: %v", v)
    case map[interface{}]interface{}:
        fmt.Printf("Map: %v", v)
        fmt.Printf("\nValue in map: %v", myInterface.(map[interface{}]interface{})["Ruben"])
    case []interface{}:
        fmt.Printf("Slice: %v", v)
    default:
        // And here I'm feeling dumb. ;)
        fmt.Printf("I don't know, ask stackoverflow.")
    }
    fmt.Println()

}
func tryInterface(i interface{}) {
    in, ok := i.(int)
    if ok {
        fmt.Println(in)
    }
}
