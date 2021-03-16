package main

import (
    "os"
    "log"
    "fmt"
    "bufio"
    "strconv"
)

func RetrieveROM(filename string) ([]byte, error) {
    file, err := os.Open(filename)

    if err != nil {
        return nil, err
    }
    defer file.Close()

    stats, statsErr := file.Stat()
    if statsErr != nil {
        return nil, statsErr
    }

    var size int64 = stats.Size()
    bytes := make([]byte, size)

    bufr := bufio.NewReader(file)
    _,err = bufr.Read(bytes)

    return bytes, err
}



const TOKEN_INTEGER = byte('i')
const TOKEN_LIST = byte('l')
const TOKEN_DICT = byte('d')
const TOKEN_END = byte('e')
const TOKEN_STRING_SEPARATOR = byte(':')

var INFO_DICT_KEY_ORDER = [4]string{"length", "name", "piece length", "pieces"}

func main() {
    data, err := RetrieveROM("ubuntu-16.04.1-server-amd64.iso.torrent")
    if err != nil {
        log.Fatal(err)
    }
    //fmt.Println(string(data)) 
    decoder := Decoder{_data: data, _index:0}
    torrent, ok := decoder.decode()
    if !ok {
        panic("EOFError! - Unexpected end of file!")
    }
    fmt.Println(torrent.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["piece length"])
    fmt.Println(torrent.(map[interface{}]interface{})["announce"])
    //fmt.Println(torrent)

    encoder := Encoder{_data:torrent.(map[interface{}]interface{})["info"]}
    fmt.Println("ENcoded info:  ", encoder.encode())
}

type Encoder struct {
    _data interface{}
}

func (e *Encoder) encode() string {
    return e.encode_next(e._data)
}

func (e *Encoder) encode_next(myInterface interface{}) string {
    switch v := myInterface.(type) {
    case int:
        // v is an int here, so e.g. v + 1 is possible.
        return e._encode_int(myInterface.(int)) + "e"
    case float64:
        // v is a float64 here, so e.g. v + 1.0 is possible.
        fmt.Printf("Float64: %v", v)
    case string:
        // v is a string here, so e.g. v + " Yeah!" is possible.
        return e._encode_string(myInterface.(string))
    case map[interface{}]interface{}:
        return e._encode_map(myInterface.(map[interface{}]interface{}))
    case []interface{}:
        return e._encode_slice(myInterface.([]interface{}))
    default:
        // And here I'm feeling dumb. ;)
        fmt.Printf("I don't know, ask stackoverflow.")
    }
    return ""
}

func (e *Encoder) _encode_int(value int) string {
    return "i" + strconv.Itoa(value)
}

func (e *Encoder) _encode_string(value string) string {
    return strconv.Itoa(len(value)) + ":" + value
}

func (e *Encoder) _encode_slice(data []interface{}) string {
    result := "l"
    for _, item := range data {
        result += e.encode_next(item)
    }
    result += "e"
    return result
}

func (e *Encoder) _encode_map(data map[interface{}]interface{}) string {
    result := "d"
    for _, item := range INFO_DICT_KEY_ORDER {
        key := e.encode_next(item)
        value := e.encode_next(data[item])
        result += key
        result += value
    }
    result += "e"
    return result
}

type Decoder struct {
    _data []byte
    _index int
}

func (d *Decoder) decode() (interface{}, bool) {
    c, ok := d._peek()
    if !ok {
        panic("EOFError! - Unexpected end of file!")
    } else if c == TOKEN_INTEGER {
        d._consume()
        return d._decode_int(), true
    } else if c == TOKEN_LIST {
        d._consume()
        return d._decode_list(), true
    } else if c == TOKEN_DICT {
        d._consume()
        return d._decode_dict(), true
    } else if c == TOKEN_END {
        return 0, false
    } else if isInByteArray(c, []byte("0123456789")) {
        // fmt.Println("---------------------> decoding integer for a string")
        return d._decode_string(), true
    } else {
        panic("Invalid token")
    }
}

func isInByteArray(c byte, bSlice []byte) bool {
    for _, value := range bSlice {
        if value == c {
            return true
        }
    }
    return false
}

func (d *Decoder) _peek() (byte, bool) {
    if d._index +1 >= len(d._data) {
        return 0, false
    }
    return d._data[d._index:d._index + 1][0], true
}

func (d *Decoder) _consume() {
    d._index += 1
}

func (d *Decoder) _read(length int) []byte {
    if d._index + length > len(d._data) {
        panic("Cannot rad bytes from current position")
    }
    res := d._data[d._index:d._index + length]
    d._index += length
    return res
}

func (d *Decoder) _read_until(token byte) []byte {
    occurrence, found := index(d._data, token, d._index)
    if !found {
        panic("Unable to find token")
    }
    result := d._data[d._index:occurrence]
    d._index = occurrence + 1
    return result
}

func (d *Decoder) _decode_int() int {
    return parseIntFromBytes(d._read_until(TOKEN_END))
}

func (d *Decoder) _decode_list() []interface{} {
    res := []interface{}{}

    for d._data[d._index: d._index + 1][0] != TOKEN_END {
        value, ok := d.decode()
        if !ok {
            log.Fatal(ok)
        }
        res = append(res, value)
    }
    d._consume() // consure the END token
    return res
}

func (d *Decoder) _decode_dict() interface{} {
    res := make(map[interface{}]interface{})
    for d._data[d._index  : d._index + 1][0] != TOKEN_END {
        key, ok := d.decode()
        obj, ok := d.decode()
        if !ok {
            log.Fatal(ok)
        }
        res[key] = obj
    }
    d._consume()  // the END token
    return res
}

func (d *Decoder) _decode_string() interface{} {
    bytes_to_read := parseIntFromBytes(d._read_until(TOKEN_STRING_SEPARATOR))
    data := d._read(bytes_to_read)
    // fmt.Printf("-----------------------> parsed string of type %T", data)
    return string(data)
}

func parseIntFromBytes(b []byte) int {
    str := string(b[:])
    i, err := strconv.Atoi(str)
    if err != nil {
        log.Fatal(err)
    }
    // fmt.Printf("------------------------->  parsed integer: %#v\n", i)
    return i
}

func index(data []byte, token byte, index int) (int, bool) {
    for i := index; i < len(data); i++ {
        if data[i] == token {
            return i, true
        }
    }
    return -1, false
}
