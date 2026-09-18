package main
import (
	"fmt"
	"sync/atomic"
)


// CircularQueue represents a lock-free SPSC circular queue
type Qdata struct{
	tr    string
	data  []byte
}
type CircularQueue struct {
	data      []Qdata
	qSize     int         // Total size of the buffer
	writePos  int32       // Next position to write 
	readPos   int32       // Next position to read 
	readCh    chan struct{}
}
var (
	WriteCount , ReadCount int
)
// NewCircularQueue initializes a new circular queue
func NewCircularQueue(size int) *CircularQueue {

	WriteCount = 0
	ReadCount = 0

	return &CircularQueue{
		data:   make([]Qdata, size),
		qSize: size,
		writePos: 0,
		readPos:  0,
		readCh : make(chan struct{}, 1),
	}
	
}

func (q *CircularQueue)Write(tr string, data []byte)bool{
	wPos := atomic.LoadInt32(&q.writePos)
	rPos := atomic.LoadInt32(&q.readPos)
	nextPos := (wPos + 1) % int32(q.qSize)
	if nextPos == rPos {
		return false
	}

	copyData := make([]byte, len(data))
	copy(copyData, data)

	current := &q.data[wPos]
	current.tr = tr
	current.data = copyData

	atomic.StoreInt32(&q.writePos, nextPos)

	WriteCount ++
	q.readCh <- struct {}{}

	return true
}

func (q *CircularQueue)Read()([]Qdata, bool){
	wPos := atomic.LoadInt32(&q.writePos)
	rPos := atomic.LoadInt32(&q.readPos)
	if rPos == wPos{
		return  nil, false
	}
	var rData []Qdata

	fmt.Printf("\tREAD Wpos(%d) Rpos(%d)\n", wPos, rPos)

	for rPos != wPos {
		rData = append(rData, q.data[rPos])
		q.data[rPos].data = nil  // Hey GC! Clean memeory
		rPos = (rPos +1 ) % int32(q.qSize)
		ReadCount ++
	}
	atomic.StoreInt32(&q.readPos, rPos)

	return rData, true
}
