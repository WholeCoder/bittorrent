package main

import (
    "fmt"
    "sort"
    "strings"
    "crypto/sha1"
)

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
    peers []BitField

}
