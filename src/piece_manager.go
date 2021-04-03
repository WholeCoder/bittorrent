package main

import (
    "os"
    "log"
    "fmt"
    "sort"
    "math"
    "time"
    "strings"
    "crypto/sha1"
)

//fd *os.File // **important**  Must call defer pieceManager.closeFile() to ensure proprer closing!!

const (
    MissingEnum = iota
    PendingEnum
    RetrievedEnum
)

type Block struct {
    piece int
    offset int
    length int
    status int
    data []byte
}

func (b *Block) init(piece int, offset int, length int) {
    b.piece = piece
    b.offset = offset
    b.length = length
    b.status = MissingEnum
    b.data = nil
}

type PieceWithBlocks struct {
   index int
   blocks []Block
   hash string
}

func (p *PieceWithBlocks) init(index int, blocks []Block, hash_value string) {
    p.index = index
    p.blocks = blocks
    p.hash = hash_value
}

func (p *PieceWithBlocks) reset() {
    for _, block := range p.blocks {
        block.status = MissingEnum
    }
}

func (p *PieceWithBlocks) next_request() *Block {
    missing := []Block{}
    for _, b := range p.blocks {
        if b.status == MissingEnum {
            missing = append(missing, b)
        }
    }
    if len(missing) > 0 {
        missing[0].status = PendingEnum
        return &missing[0]
    }
    return nil
}

func (p *PieceWithBlocks) block_received(offset int, data []byte) {
    matches := []Block{}
    for _, b := range p.blocks {
        if b.offset == offset {
            matches = append(matches,b)
        }
    }
    var block *Block = nil
    if len(matches) >0 {
        block = &matches[0]
    }
    if block != nil {
        block.status = RetrievedEnum
        block.data = data
    } else {
        fmt.Printf("\nTyring to complete a non-existing block %v", offset)
    }
}

func (p *PieceWithBlocks) is_complete() bool {
    blocks := []Block{}
    for _, b := range p.blocks {
        if b.status != RetrievedEnum {
            blocks = append(blocks, b)
        }
    }
    return len(blocks) == 0
}

func (p *PieceWithBlocks) is_hash_matching() bool {
    h := sha1.New()
    h.Write([]byte(p.data()))
    piece_hash := string(h.Sum(nil))

    return p.hash == piece_hash
}

func (p *PieceWithBlocks) data() string {
    sort.SliceStable(p.blocks, func(i, j int) bool {
        return p.blocks[i].offset < p.blocks[j].offset
    })
    blocks_data := []string{}
    for _, b := range p.blocks {
        blocks_data = append(blocks_data, string(b.data))
    }

    return strings.Join(blocks_data,"")
}

type PendingRequest struct {
    block Block
    added int
}

type PieceManager struct {
    torrent Torrent
    peers map[string]BitsetByte
    pending_blocks []PendingRequest
    missing_pieces []PieceWithBlocks
    ongoing_pieces []PieceWithBlocks
    have_pieces []PieceWithBlocks
    max_pending_time int
    total_pieces int
    fd *os.File // **important**  Must call defer pieceManager.Close() to ensure proprer closing!!
}

func (p *PieceManager) init(torrent Torrent) {
    p.torrent = torrent
    // p.peers = {}
    // p.pending_blocks = []
    // p.missing_pieces = []
    // p.ongoing_pieces = []
    // p.have_pieces = []
    p.max_pending_time = 300 * 1000 // 5 minutes
    p.missing_pieces = p._initiate_pieces()
    p.total_pieces = len(torrent.pieces())
    fd, err := os.Create(p.torrent.output_file())
    if err != nil {
        log.Fatal(err)
    }
    p.fd = fd
}

const REQUEST_SIZE float64 = 16384.0 // 2^14

func (p *PieceManager) _initiate_pieces() []PieceWithBlocks {
    torrent := p.torrent
    pieces  := []PieceWithBlocks{}
    total_pieces := len(torrent.pieces())
    std_piece_blocks := int(math.Ceil(float64(torrent.piece_length()) / REQUEST_SIZE))
    for index, hash_value := range torrent.pieces() {
        blocks := []Block{}
        if index < (total_pieces - 1) {
            for offset := 0; offset < std_piece_blocks; offset++ {
                b := Block{}
                b.init(index, offset * int(REQUEST_SIZE), int(REQUEST_SIZE))
                blocks = append(blocks, b)
            }
        } else {
            last_length := torrent.total_size() % torrent.piece_length()
            num_blocks := int(math.Ceil(float64(last_length) / REQUEST_SIZE))
            for offset := 0; offset < num_blocks; offset++ {
                b := Block{}
                b.init(index, offset * int(REQUEST_SIZE), int(REQUEST_SIZE))
                blocks = append(blocks, b)
            }
            if last_length % int(REQUEST_SIZE) > 0 {
                last_block := blocks[len(blocks)-1]
                last_block.length = last_length % int(REQUEST_SIZE)
                blocks[len(blocks)-1] = last_block
            }
        }
        p := PieceWithBlocks{}
        p.init(index, blocks, hash_value)
        pieces = append(pieces, p)
    }

    return pieces
}

// *** important **  Must call defer pieceManager.closeFile
func (p *PieceManager) closeFile() {
    p.fd.Close()
}

func (p *PieceManager) complete() bool {
    return len(p.have_pieces) == p.total_pieces
}

func (p *PieceManager) bytes_dowloaded() int {
    return len(p.have_pieces) * p.torrent.piece_length()
}

func (p *PieceManager) bytes_uploaded() int {
    return 0
}

func (p *PieceManager) add_peer(peer_id string, bitfield BitsetByte) {
    p.peers[peer_id] = bitfield
}

func (p *PieceManager) update_peer(peer_id string, index int) {
    _, ok := p.peers[peer_id]
    if ok {
        p.peers[peer_id].SetBit(index, true)
    }
}

func (p *PieceManager) remove_peer(peer_id string) {
    _, ok := p.peers[peer_id]
    if ok {
        delete(p.peers, peer_id)
    }
}

func (p *PieceManager) next_request(peer_id string) Block {
    _, ok := p.peers[peer_id]
    if !ok {
        return nil
    }

    block = p._expired_requests(peer_id)
    if block == nil {
        block = p._next_ongoing(peer_id)
        if block == nil {
            block = p._get_rarest_piece(peer_id).next_request()
        }
    }
    return block
}

func RemovePendingRequestIndex(s []PendingRequest, index int) []PendingRequest {
    ret := make([]PendingRequest, 0)
    ret = append(ret, s[:index]...)
    return append(ret, s[index+1:]...)
}

func RemovePieceWithBlocks(s []PieceWithBlocks, pWithBlocks PieceWithBlocks) []PieceWithBlocks {
    ret := make([]PieceWithBlocks, 0)
    index := 0
    for s[index] != pWithBlocks {
        index++
    }
    ret = append(ret, s[:index]...)
    return append(ret, s[index+1:]...)
}

func (p *PieceManager) block_received(peer_id string, piece_index int, block_offset int, data []byte) {
    fmt.Printf("Received block %d for piece %d from peer %s: ", block_offset, piece_index, peer_id)

    for index, request := range p.pending_blocks {
        if request.block.piece == piece_index {
            p.pending_blocks = RemovePendingRequestIndex(p.pending_blocks, index)
            break
        }
    }
    pieces := []PieceWithBlocks{}
    for _, p := range p.ongoing_pieces {
        if p.index == piece_index {
            pieces = append(pieces, p)
        }
    }
    var piece Piece = nil
    if len(pieces) > 0 {
        piece = pieces[0]
    }
    if piece != nil {
        piece.block_received(block_offset, data)
        if piece.is_complete() {
            if piece.is_hash_matching() {
                p._write(piece)
                RemovePieceWithBlocks(p.ongoing_pieces, piece)
                p.have_pieces = append(p.have_pieces, piece)
                complete := (p.total_pieces - len(p.missing_pieces) - len(p.ongoing_pieces))
                fmt.Printf("\n%d / %d pieces downloaded %.3f %%\n",complete, p.total_pieces, float64(complete)/float64(p.total_pieces)*100.0)
            } else {
                fmt.Printf("\nDiscarding corrupt piect %d", piece.index)
                piece.reset()
            }
        }
    } else {
        fmt.Println("\nTrying to update piece that is not ongoing!")
    }
}

func (p *PieceManager) _expired_requests(peer_id string) Block {
    current := time.Now().Unix() * 1000 // milli-seconds since epoch (int)
    for _, request := range p.pending_blocks {
        if p.peers[peer_id].GetBit(request.block.piece) {
            if request.added + p.max_pending_time < current {
                fmt.Printf("\nRe-requesting block %d for piece %d\n", request.block.offset, request.block.piece)
                request.added = current
                return request.block
            } // end if
        } // if
    } // end for request := p.pending_blocks
    return nil
}

func (p *PieceManager) _next_ongoing(peer_id string) Block {
    for _, piece := range p.ongoing_pieces {
        if p.peers[peer_id].GetBit(piece.index) {
            block = piece.next_request()
            if block != nil {
                p.pending_blocks = append(p.pending_blocks, PendingRequest{block: block, added: time.Now().Unix() * 1000})
                return block
            } /// if block != nil
        } // end if GetBit
    } // for range p.ongoing_pieces
    return nil
}

func (p *PieceManager) _get_rarest_piece(peer_id string) PieceWithBlocks {
    piece_count := map[Piece]int{}
    for _, piece := range p.missing_pieces {
        if !p.peers[peer_id].GetBit(piece.index) {
            continue
        }
        for _, p := range p.peers {
            if p.peers[p].GetBit(piece.index) {
                piece_count[piece] += 1
            } // end GetBit
        } // end for p.peers
    } // end for missing_pieces
    rarest_piece := 
}
