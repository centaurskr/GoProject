package main
import (
	"fmt"
	//"sync/atomic"
	//"time"
)
type CircularQueue struct {
	buffer    []byte
	qSize     int
	writePos  int
	readPos   int
	notifyCh  chan struct{}  // Notify reader when data is available
	spaceCh   chan struct{}  // Notify writer when space is available
}


// NewCircularQueue initializes a new circular queue
func NewCircularQueue(size int) *CircularQueue {
	q := &CircularQueue{
		buffer:   make([]byte, size),
		qSize:    size,
		notifyCh: make(chan struct{}, 1),
		spaceCh:  make(chan struct{}, 1),
	}
	q.spaceCh <- struct{}{} // Initially, space is available
	return q
}

// FreeSpace calculates available space in the queue
func (q *CircularQueue) FreeSpace() int {
	if q.writePos >= q.readPos {
		return q.qSize - (q.writePos - q.readPos) - 1
	}
	return (q.readPos - q.writePos) - 1
}

// Write adds data to the queue
func (q *CircularQueue) Write(tr string, data []byte) bool {
	trByte := []byte(tr)
	trLen := len(trByte)
	dataLen := len(data)
	recordLen := 4 + trLen + dataLen

	// Check for available space
	if q.FreeSpace() < recordLen+4 {
		return false
	}

	// Write Record Size
	q.buffer[q.writePos] = byte((recordLen >> 24) & 0xFF)
	q.buffer[(q.writePos+1)%q.qSize] = byte((recordLen >> 16) & 0xFF)
	q.buffer[(q.writePos+2)%q.qSize] = byte((recordLen >> 8) & 0xFF)
	q.buffer[(q.writePos+3)%q.qSize] = byte((recordLen) & 0xFF)
	q.writePos = (q.writePos + 4) % q.qSize

	// Write trLen
	q.buffer[q.writePos] = byte((trLen >> 24) & 0xFF)
	q.buffer[(q.writePos+1)%q.qSize] = byte((trLen >> 16) & 0xFF)
	q.buffer[(q.writePos+2)%q.qSize] = byte((trLen >> 8) & 0xFF)
	q.buffer[(q.writePos+3)%q.qSize] = byte((trLen) & 0xFF)
	q.writePos = (q.writePos + 4) % q.qSize

	// Write tr
	for i := 0; i < trLen; i++ {
		q.buffer[q.writePos] = trByte[i]
		q.writePos = (q.writePos + 1) % q.qSize
	}

	// Write data
	for i := 0; i < dataLen; i++ {
		q.buffer[q.writePos] = data[i]
		q.writePos = (q.writePos + 1) % q.qSize
	}

	// Notify reader that data is available
	select {
	case q.notifyCh <- struct{}{}:
	default:
	}

	return true
}

// Read retrieves data from the queue
func (q *CircularQueue) Read() (string, []byte, bool) {
	if q.readPos == q.writePos {
		return "", nil, false
	}

	// Read Record Size
	recordLen := int(q.buffer[q.readPos])<<24 |
		int(q.buffer[(q.readPos+1)%q.qSize])<<16 |
		int(q.buffer[(q.readPos+2)%q.qSize])<<8 |
		int(q.buffer[(q.readPos+3)%q.qSize])
	q.readPos = (q.readPos + 4) % q.qSize

	if recordLen <= 4 {
		fmt.Println("Error: Corrupted record length", recordLen)
		return "", nil, false
	}

	// Read trLen
	trLen := int(q.buffer[q.readPos])<<24 |
		int(q.buffer[(q.readPos+1)%q.qSize])<<16 |
		int(q.buffer[(q.readPos+2)%q.qSize])<<8 |
		int(q.buffer[(q.readPos+3)%q.qSize])
	q.readPos = (q.readPos + 4) % q.qSize

	// Read tr
	tr := make([]byte, trLen)
	for i := 0; i < trLen; i++ {
		tr[i] = q.buffer[q.readPos]
		q.readPos = (q.readPos + 1) % q.qSize
	}

	// Read data
	dataLen := recordLen - (4 + trLen)
	data := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		data[i] = q.buffer[q.readPos]
		q.readPos = (q.readPos + 1) % q.qSize
	}

	// Notify writer that space is available
	select {
	case q.spaceCh <- struct{}{}:
	default:
	}

	return string(tr), data, true
}
