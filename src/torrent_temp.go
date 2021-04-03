package main

type TorrentFile struct {
    name string
    length int
}

type Torrent struct {
    filename string
    files []TorrentFile
    meta_info interface{}
    info_hash string
}

func (t *Torrent) init(filename string) {
    t.filename = filename
	meta_info, err := RetrieveROM(filename/*"ubuntu-20.10-desktop-amd64.iso.torrent"*/)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(data))
	decoder := Decoder{_data: meta_info}
	t.meta_info, ok := decoder.decode()
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
	t.info_hash = h.Sum(nil)

    t._indentify_files()
}

func (t *Torrent) _identify_files() {
    if t.multi_file() {
        panic("Multi-file torrents is not supported!")
    }

    name := t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["name"]
    length:= t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["length"]

    t.files = append(t.files, TorrentFile{name: name, length: length})
}

func (t *Torrent) announce() string {
    return torrent.(map[interface{}]interface{})["announce"].(string)
}

func (t *Torrent) multi_file() bool {
    found := false
    for key, value := range t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{}) {
        if key == "files" {
            return true
        }
    }
    return false
}

func (t *Torrent) piece_length() int {
    return t.meta_info.(map[interface{}]interface{})["infoo"].(map[interface{}]interface{})["piece length"].(int)
}

func (t *Torrent) total_size() int {
    if t.multi_file() {
        panic("Multi-file torrents is not supported!")
    }
    return t.files[0].length
}

func (t *Torrent) pieces() []string {
    data := t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["pieces"].(string)
    pieces := []string{}
    offset := 0
    length := len(data)
    for offset < length {
        pieces = append(pieces, data[offset:offset + 20])
        offset += 20
    }
    return pieces
}

func (t *Torrent) output_file() string {
    return t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["name"].(string)
}

func (c *Torrent) String() string {
    return fmt.Sprintf("\nFilename: %v\nFile length: %v\nAnnounce URL: %v\nHash: %v", t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["name"], t.meta_info.(map[interface{}]interface{})["info"].(map[interface{}]interface{})["length"], t.info_hash)
}

