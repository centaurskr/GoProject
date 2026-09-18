package main

import (
    "fmt"
    "math/rand"
    "time"
)
// Updated struct with Payload as [][]byte
type Packet struct {
    ID      int
    Payload [][]byte // Now a slice of byte slices
    Length  int      // Total number of sub-slices
    Stamp   time.Time
}

var Channel chan Packet

func goProcess() {
    for {
        select {
        case pkt := <- Channel:
            fmt.Printf("ID[%d] Chunks[%d] Timestamp[%v] Data[%v]\n", pkt.ID, pkt.Length, pkt.Stamp, pkt.Payload)
        }
    }
}

func main() {
    Channel = make(chan Packet, 10000)
    go goProcess()

    rand.Seed(time.Now().UnixNano())

    i := 0
    for {
        numChunks := rand.Intn(5) + 1 // Random number of chunks (1-5)
        var data [][]byte

        for c := 0; c < numChunks; c++ {
            chunkLen := rand.Intn(10) + 1 // Each chunk has 1-10 bytes
            chunk := make([]byte, chunkLen)
            for j := range chunk {
                chunk[j] = byte(rand.Intn(256)) // Fill with random bytes
            }
            data = append(data, chunk)
        }

        pkt := Packet{
            ID:      i,
            Payload: data,
            Length:  len(data),
            Stamp:   time.Now(),
        }
        Channel <- pkt
        i++
        time.Sleep(100 * time.Millisecond) // Avoid excessive CPU usage
    }
}

