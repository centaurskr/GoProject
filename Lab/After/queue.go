///
/// @brief  : Circula Queue 
/// @file   : circulaQ.go
/// @date   : 2025. 04. 01. (화) 12:19:57 KST
/// @author : Cento
///
package main
import (
	//"sync/atomic"
	"fmt"
)
var (
	WriteCount int
	ReadCount  int
)
// CircularQueue represents a lock-free SPSC circular queue
type CircularQueue struct {
	buffer   []byte      // Fixed-size buffer
	qSize     int         // Total size of the buffer
	writePos int         // Next position to write 
	readPos  int         // Next position to read 
}

// NewCircularQueue initializes a new circular queue
func NewCircularQueue(size int) *CircularQueue {

	WriteCount = 0
	ReadCount = 0

	return &CircularQueue{
		buffer:   make([]byte, size),
		qSize: size,
		writePos: 0,
		readPos:  0,
	}
}

// FreeSpace calculates available space in the queue
func (q *CircularQueue) FreeSpace() int {
	if q.writePos >= q.readPos{
		return q.qSize - (q.writePos - q.readPos) -1
	}
	return (q.readPos - q.writePos ) -1
}

func (q *CircularQueue) Write(tr string, data []byte) bool {
	dataLen   := len(data)
	trByte    := []byte(tr)
	trLen     := len(trByte)
	recordLen := 4 + trLen + dataLen

	if q.FreeSpace() < (recordLen+4){
		fmt.Printf("Queue Full : Cannot Write. Wpos(%d) Rpos(%d). Rlen(%d)\n", 
			q.writePos, q.readPos, recordLen)
		return false
	}
	
	// Write Record Size
	q.buffer[ q.writePos     ] = byte(recordLen >> 24)
	q.buffer[ (q.writePos + 1) % q.qSize ] = byte(recordLen >> 16)
	q.buffer[ (q.writePos + 2) % q.qSize ] = byte(recordLen >> 8 )
	q.buffer[ (q.writePos + 3) % q.qSize ] = byte(recordLen)
	q.writePos = (q.writePos + 4 ) % q.qSize

	// Write trLen
	q.buffer[ q.writePos     ] = byte(trLen >> 24)
	q.buffer[ (q.writePos + 1) % q.qSize ] = byte(trLen >> 16)
	q.buffer[ (q.writePos + 2) % q.qSize ] = byte(trLen >> 8 )
	q.buffer[ (q.writePos + 3) % q.qSize ] = byte(trLen)
	q.writePos = (q.writePos + 4 ) % q.qSize

	// Copy Tr 
	for i:= 0; i < trLen;i++{
		q.buffer[ q.writePos ] = trByte[i]
		q.writePos = (q.writePos + 1) % q.qSize
	}

	// Copy data
	for i:= 0; i < dataLen;i++{
		q.buffer[ q.writePos ] = data[i]
		q.writePos = (q.writePos + 1) % q.qSize
	}

	WriteCount ++

	fmt.Printf("WRITE q.W(%d) q.R(%d) - Rl(%d) W/R(%d/%d)\n", 
		q.writePos, q.readPos, recordLen, WriteCount, ReadCount) 


	return true
}

// Read retrieves the next record and splits it into trString and data
func (q *CircularQueue) Read() (string, []byte, bool) {
	if q.readPos == q.writePos {
		return "", nil, false
	}

	recordLen := int(q.buffer[ q.readPos ]) << 24 |
				int(q.buffer[ (q.readPos + 1) % q.qSize ] ) << 16 |
				int(q.buffer[ (q.readPos + 2) % q.qSize ] ) << 8  |
				int(q.buffer[ (q.readPos + 3) % q.qSize ] )
	q.readPos = (q.readPos + 4) % q.qSize

    // Read trLen
    trLen := int(q.buffer[q.readPos])<<24 |
        int(q.buffer[(q.readPos+1)%q.qSize])<<16 |
        int(q.buffer[(q.readPos+2)%q.qSize])<<8 |
        int(q.buffer[(q.readPos+3)%q.qSize])
    q.readPos = (q.readPos + 4) % q.qSize

    // Extract tr (topic)
    tr := make([]byte, trLen)
    for i := 0; i < trLen; i++ {
        tr[i] = q.buffer[q.readPos]
        q.readPos = (q.readPos + 1) % q.qSize
    }

    // Extract data
    dataLen := recordLen - (4 + trLen)
    data := make([]byte, dataLen)
    for i := 0; i < dataLen; i++ {
        data[i] = q.buffer[q.readPos]
        q.readPos = (q.readPos + 1) % q.qSize
    }

/******
	record := make([]byte, recordLen)
	for i := 0 ; i < recordLen; i++{
		record[i] = q.buffer[ q.readPos ]
		q.readPos = (q.readPos + 1) % q.qSize
	}
*************/


	ReadCount ++

	fmt.Printf("\tREAD q.W(%d) q.R(%d) - Rl(%d) W/R(%d/%d)\n", 
		q.writePos, q.readPos, recordLen, WriteCount, ReadCount) 

	return string(tr), data, true
}
