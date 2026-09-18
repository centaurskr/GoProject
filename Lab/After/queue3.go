package main
import (
	"fmt"
)


// CircularQueue represents a lock-free SPSC circular queue
type Qdata struct{
	tr    string
	data  []byte
}
type CircularQueue struct {
	data      []Qdata
	qSize     int         // Total size of the buffer
	writePos int         // Next position to write 
	readPos  int         // Next position to read 
	dataCh  chan struct{}
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
		dataCh : make(chan struct{}, 1),
	}
}

func (q *CircularQueue)Write(tr string, data []byte)bool{
	nextPos := (q.writePos + 1) % q.qSize
	if nextPos == q.readPos {
		return false
	}

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	current := &q.data[q.writePos]
	current.tr = tr
	current.data = dataCopy

	q.writePos = nextPos

	WriteCount ++
	q.dataCh <- struct {}{}

	return true
}

func (q *CircularQueue)Read()(string, []byte, bool){
	if q.readPos == q.writePos{
		return "", nil, false
	}

	current := &q.data[q.readPos]
	tr := current.tr
	data := current.data

	// Clear the reference to allow GC
	current.data = nil
	q.readPos = (q.readPos + 1) % q.qSize

	ReadCount ++
	fmt.Printf("\tREAD Wpos(%d) Rpos(%d) W/R (%d:%d)\n", q.writePos, q.readPos, WriteCount, ReadCount)
	return tr, data, true
}
