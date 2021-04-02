package main

import (
    "fmt"
)



type TrackerResponse struct {
    response interface{}
}

func (t *TrackerResponse) init(response interface{}) {
    t.response = response
}

func (t *TrackerResponse) failure() string {
    value, ok := t.response.(map[interface{}]interface{})["failure reason"]
    if ok {
        return value.(string)
    }
    return nil
}

func (t *TrackerResponse) interval() int {
    value, ok := t.response.(map[interface{}]interface{})["interval"]
    if ok {
        return value.(int)
    }
    return 0
}

func (t *TrackerResponse) complete() int {
    value, ok := t.response.(map[interface{}]interface{})["complete"]
    if ok {
        return value.(int)
    }
    return 0
}

func (t *TrackerResponse) incomplete() int {
    value, ok := t.response.(map[interface{}]interface{})["incomplete"]
    if ok {
        return value.(int)
    }
    return 0
}

func (t *TrackerResponse) peers() []PeerStruct{
    peers := t.response.(map[interface{}]interface{})["peers"]
    switch v := peers.(type) {
    case []interface{}:
        panic("Dictionary model peers are returned by tracker. - NotImplementedError.")
    }
    fmt.Println("Binary model peers are returned by tracker")

    peersString = peers.(string)

    peersByteSlice := []string{}
    for i := 0; i < len(peersString); I += 6 {
        peersByteSlice = append(peersByteSlice, peersString[i:i+6])
    }

    peersStructSlice := []PeerStruct{}
    for _, value := range peersByteSlice {
        peersStructSlice = append(peersStructSlice, PeerStruct{address: value[:4], port: value[4:]})
    }

    return peersStructSlice
}

func (t *TrackerResponse) String() string {
    return fmt.Sprintf("\nincoplete: %v\ncomplete: %v\ninterval: %v\npeers: %v\n", t.incomplete(), t.complete(), t.interval(), t.peers)
}

type PeerStruct struct {
    address string
    port string
}
