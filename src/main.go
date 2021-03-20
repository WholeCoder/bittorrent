package main

import (
	"fmt"
	"log"
	"net"
	"os"
    "io"
	//"math"
	"bufio"
    "bytes"
	"crypto/sha1"
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
    "github.com/lunixbochs/struc"
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
	_, err = bufr.Read(bytes)

	return bytes, err
}

const TOKEN_INTEGER = byte('i')
const TOKEN_LIST = byte('l')
const TOKEN_DICT = byte('d')
const TOKEN_END = byte('e')
const TOKEN_STRING_SEPARATOR = byte(':')

var INFO_DICT_KEY_ORDER = [4]string{"length", "name", "piece length", "pieces"}

func main() {
	// data, err := RetrieveROM("ubuntu-16.04.1-server-amd64.iso.torrent")
	data, err := RetrieveROM("ubuntu-20.10-desktop-amd64.iso.torrent")
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(data))
	decoder := Decoder{_data: data, _index: 0}
	torrent, ok := decoder.decode()
	if !ok {
		panic("EOFError! - Unexpected end of file!")
	}
	fmt.Println(torrent.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["piece length"])
	fmt.Println(torrent.(map[interface{}]interface{})["announce"])
	//fmt.Println(torrent)

	encoder := Encoder{_data: torrent.(map[interface{}]interface{})["info"]}
	// fmt.Println("ENcoded info:  ", encoder.encode())
	info_hash := encoder.encode()
	h := sha1.New()
	h.Write([]byte(info_hash))
	bs := h.Sum(nil)

	params := make(url.Values)
	params.Add("info_hash", string(bs))

	peer_id := "-PC0001-"
	for i := 0; i < 12; i++ {
		peer_id += string(rand.Intn(10))
	}
	params.Add("peer_id", peer_id)
	params.Add("port", strconv.Itoa(6889))
	params.Add("uploaded", strconv.Itoa(0))
	params.Add("downloaded", strconv.Itoa(0))
	params.Add("left", strconv.Itoa(torrent.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["length"].(int)))
	params.Add("compact", strconv.Itoa(1))

	first := true
	if first {
		params.Add("event", "started")
	}
	url := torrent.(map[interface{}]interface{})["announce"].(string) + "?" + params.Encode()
	// fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("data retrieved: ", string(b))

	peerDecoder := Decoder{_data: b, _index: 0}
	peerMap, ok := peerDecoder.decode()
	if !ok {
		panic("EOFError! - Unexpected end of file!")
	}
	peers := peerMap.(map[interface{}]interface{})["peers"].(string)

    ipBytes := peers[0:4]
	ip := net.IP(ipBytes)

	port := binary.BigEndian.Uint16([]byte(peers[4:6]))

    service := ip.String() + ":" + strconv.Itoa(int(port))
    fmt.Println("Service: ", service)

    tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
    if err != nil {
        log.Fatal(err)
    }

    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    hShake := Handshake{}
    hShake.init(bs, peer_id)

    _, err = conn.Write(hShake.encode())
    if err != nil {
        log.Fatal(err)
    }
	// read 2^14 bytes from the Reader called r
	n := 68
	p := make([]byte, n)
	_, err = io.ReadFull(conn, p)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Read in bytes as string: %v\n", string(p))
    fmt.Printf("Read in bytes: %#v\nLength of bytes: %v\n", p, len(p))
    fmt.Printf("\nStruct Returned = %#v\n",hShake.decode(p))
    returnHShake := hShake.decode(p).(*Handshake)
    fmt.Printf("Length of peer_id: %v\n", len(returnHShake.peer_id))
    fmt.Printf("Length of info_hash: %v\n", len(returnHShake.info_hash))


} // main

// PeerMessage Enums
const (
	ChokeEnum = iota
	UnchokeEnum
	InterestedEnum
	NotInterestedEnum
	HaveEnum
	BitFieldEnum
	RequestEnum
	PieceEnum
	CancelEnum
	PortEnum
	HandshakeEnum // listed as None in python client
	KeepAliveEnum // listed as None in python client
)

type PeerMessage interface {
	encode() []byte
	decode([]byte) PeerMessage
}

type HandshakeStruct struct {
    Length byte `struc:"big"`
    Protocol string `struc:"[19]byte"`
    Space [8]byte
    Info_hash string `struc:"[20]byte"`
    Peer_id string `struc:"[20]byte"`
}


type Handshake struct {
	info_hash string
	peer_id   string
}

// Handshake Constructor
func (h *Handshake) init(info_hash []byte, peer_id string) {
	h.info_hash = string(info_hash)
	h.peer_id = peer_id
}

func (h *Handshake) encode() []byte {
    var buf bytes.Buffer
    t := &HandshakeStruct{byte(19), "BitTorrent protocol", [8]byte{0,0,0,0,0,0,0,0},h.info_hash, h.peer_id}
    err := struc.Pack(&buf, t)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("\n%#v\n\n", t)
    fmt.Printf("%#v\n\n", buf.Bytes())

	return buf.Bytes()
}

func (h *Handshake) decode(data []byte) PeerMessage {
	/*if len(data) < (49 + 19) {
		return nil
	}*/
    var buf bytes.Buffer
    buf.Write(data)

    o := &HandshakeStruct{}
    err := struc.Unpack(&buf, o)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Unpacked struc:  %#v\n\n", o)
    fmt.Printf("\nProtocol %v\n", o.Protocol)


	hShake := Handshake{info_hash: o.Info_hash, peer_id: o.Peer_id}

	var hShakePeerMessage PeerMessage = &hShake
	return hShakePeerMessage
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
	_data  []byte
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
	if d._index+1 >= len(d._data) {
		return 0, false
	}
	return d._data[d._index : d._index+1][0], true
}

func (d *Decoder) _consume() {
	d._index += 1
}

func (d *Decoder) _read(length int) []byte {
	if d._index+length > len(d._data) {
		panic("Cannot read bytes from current position")
	}
	res := d._data[d._index : d._index+length]
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

	for d._data[d._index : d._index+1][0] != TOKEN_END {
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
	for d._data[d._index : d._index+1][0] != TOKEN_END {
		key, ok := d.decode()
		obj, ok := d.decode()
		if !ok {
			log.Fatal(ok)
		}
		res[key] = obj
	}
	d._consume() // the END token
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
