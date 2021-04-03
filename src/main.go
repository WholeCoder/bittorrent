package main
//fd *os.File // **important**  Must call defer pieceManager.closeFile() to ensure proprer closing!!
import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"math"
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"github.com/lunixbochs/struc"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
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
	decoder := Decoder{_data: data}
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
	peerMessage := make(chan PeerMessage)

	go PeerStreamIterator(service, peerMessage, peer_id, bs)

	for message := range peerMessage {
		fmt.Printf("\nGot Message:  %#v\n", message)
	}
} // main

func PeerStreamIterator(service string, peerMessage chan PeerMessage, peer_id string, info_hash []byte) {
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
	hShake.init(info_hash, peer_id)

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
	fmt.Printf("\nStruct Returned = %#v\n", hShake.decode(p))
	returnHShake := hShake.decode(p).(*Handshake)
	fmt.Printf("Length of peer_id: %v\n", len(returnHShake.peer_id))
	fmt.Printf("Length of info_hash: %v\n", len(returnHShake.info_hash))

	remote_id := returnHShake.peer_id
	fmt.Println("\nRemote id: ", remote_id)

	message := Interested{}
	fmt.Println("Sending Message:  Interested")
	_, err = conn.Write(message.encode())
	if err != nil {
		fmt.Println("ERROR writing Interested Message")
		log.Fatal(err)
	}
	fmt.Printf("\nInterested.encode():  %#v\n", message.encode())
	n = 5
	p = make([]byte, n)
	_, err = io.ReadFull(conn, p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nBitArray lenth: %v\tMessage Id: %#v\n", int(binary.BigEndian.Uint32(p[0:4]))-1, p[4])

	p = make([]byte, binary.BigEndian.Uint32(p[0:4])-1)
	_, err = io.ReadFull(conn, p)
	if err != nil {
		log.Fatal(err)
	}

    bField := BitField{}
    bField.init(p)

	var pMessage PeerMessage = &bField
	peerMessage <- pMessage

    n = 5
	p = make([]byte, n)
	_, err = io.ReadFull(conn, p)
	if err != nil {
		log.Fatal(err)
	}
    if p[4] == UnchokeEnum {
        unchokeMsg := Unchoke{}
        pMessage = &unchokeMsg
        peerMessage <- pMessage
    }




    for i := 0; i < 2; i++ {

        requestMessage := Request{index:0, begin:i, length:int(math.Pow(2, 14))}
	    fmt.Println("Sending Message: Request")
	    _, err = conn.Write(requestMessage.encode())
	    if err != nil {
		    fmt.Println("ERROR writing Request Message")
		    log.Fatal(err)
	    }
	    fmt.Printf("\nRequest.encode():  %#v\n", requestMessage.encode())



        var buf bytes.Buffer
        n = 4
	    p = make([]byte, n)
	    _, err = io.ReadFull(conn, p)
	    if err != nil {
		    log.Fatal(err)
	    }
        buf.Write(p)

        message_length := binary.BigEndian.Uint32(p[0:4])
        fmt.Printf("\nmessage Length:  %v\n", message_length)
        n = int(message_length)
	    p = make([]byte, n)
	    _, err = io.ReadFull(conn, p)
	    if err != nil {
		    log.Fatal(err)
	    }

        buf.Write(p)

        piece := Piece{}
	    returnPiece := piece.decode(buf.Bytes()).(*Piece)

        pMessage = returnPiece
        peerMessage <- pMessage
    }
} // PeerStreamIterator

type BitsetByte []byte

func InitNewByteset(bray []byte) BitsetByte {
    return BitsetByte(bray)
}

/*func NewUint8(n int) BitsetUint8 {
	return make(BitsetUint8, (n+7)/8)
}*/

func (b BitsetByte) GetBit(index int) bool {
	pos := index / 8
	j := uint(7 - index % 8)
    fmt.Printf("\nbit at element (GetBit):  %v", b[pos])
	return (b[pos] & (byte(1) << j)) != 0
}

func (b BitsetByte) SetBit(index int, value bool) {
	pos := index / 8
    j := uint(7 - index % 8)
    fmt.Println("j =", j)
	if value {
		b[pos] |= (byte(1) << j)
	} else {
		b[pos] &= ^(byte(1) << j)
	}
    fmt.Printf("\nBitset element (in SetBit):  %8b\n", b[pos])
}

func (b BitsetByte) Len() int {
	return 8 * len(b)
}

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

type PieceStruct struct {
    Length int  `struc:"big"`
    Message_id byte
    Index int
    Begin int
}

type Piece struct {
    index int
    begin int
    block []byte
}

func (p *Piece) encode() []byte {
	// <length prefix><message ID><index><begin><block>
    var buf bytes.Buffer
	t := &PieceStruct{len(p.block) + 9, PieceEnum, p.index, p.begin}
	err := struc.Pack(&buf, t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n%#v\n\n", t)
	fmt.Printf("%#v\n\n", buf.Bytes())

	buf.Write(p.block)

    return buf.Bytes()
}

func (p *Piece) decode(data []byte) PeerMessage {
	// <length prefix><message ID><index><begin><block>
	var buf bytes.Buffer
    buf.Write(data[0:13])

	o := &PieceStruct{}
	err := struc.Unpack(&buf, o)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Unpacked struc:  %#v\n\n", o)

    piece := Piece{index: o.Index, begin: o.Begin, block: data[13:]}

	var piecePeerMessage PeerMessage = &piece
	return piecePeerMessage
}

type RequestStruct struct {
	Message_length    int   `struc:"big"`
    Message_id byte
    Index int
    Begin int
    Request_length int
}

type Request struct {
    index int
    begin int
    length int
}


func (r *Request) encode() []byte {
    // <len=0013><id=6><index><begin><length>
	var buf bytes.Buffer
	t := &RequestStruct{13, RequestEnum, r.index, r.begin, r.length}
	err := struc.Pack(&buf, t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n%#v\n\n", t)
	fmt.Printf("%#v\n\n", buf.Bytes())

	return buf.Bytes()
}

func (r *Request) decode(data []byte) PeerMessage {
	var buf bytes.Buffer
	buf.Write(data)

	o := &RequestStruct{}
	err := struc.Unpack(&buf, o)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Unpacked struc:  %#v\n\n", o)

    request := Request{index: o.Index, begin: o.Begin, length: o.Message_length}

	var requestPeerMessage PeerMessage = &request
	return requestPeerMessage
}

type BitField struct {
    bitfield BitsetByte
}

func (b *BitField) init(data []byte) {
    b.bitfield = InitNewByteset(data)
}

func (b *BitField) encode() []byte {
	var buf bytes.Buffer
	t := &LengthIdStruct{uint32(len(b.bitfield) + 1), BitFieldEnum}
	err := struc.Pack(&buf, t)
	if err != nil {
		log.Fatal(err)
	}
    buf.Write(b.bitfield)
    return buf.Bytes()
}

func (b *BitField) decode(slice []byte) PeerMessage {
    //message_length := binary.LittleEndian.Uint32(slice[0:4])
    bField := BitField{}
    bField.init(slice[5:])
	var bFieldPeerMessage PeerMessage = &bField
    return bFieldPeerMessage
}

type HandshakeStruct struct {
	Length    byte   `struc:"big"`
	Protocol  string `struc:"[19]byte"`
	Space     [8]byte
	Info_hash string `struc:"[20]byte"`
	Peer_id   string `struc:"[20]byte"`
}

type LengthIdStruct struct {
	Length uint32 `struc:"big"`
	Id     byte
}

type Interested struct {
}

type Unchoke struct {
}

func (u *Unchoke) encode() []byte {
    return nil
}

func (u *Unchoke) decode(data []byte) PeerMessage {
    return nil
}

func (i *Interested) encode() []byte {
	var buf bytes.Buffer
	t := &LengthIdStruct{1, InterestedEnum}
	err := struc.Pack(&buf, t)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func (i *Interested) decode(data []byte) PeerMessage {
	return nil
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
	t := &HandshakeStruct{byte(19), "BitTorrent protocol", [8]byte{0, 0, 0, 0, 0, 0, 0, 0}, h.info_hash, h.peer_id}
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
