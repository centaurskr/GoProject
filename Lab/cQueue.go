package main

import (
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

// CircularQueue represents a lock-free SPSC circular queue
type CircularQueue struct {
	buffer   []byte      // Fixed-size buffer
	capacity int         // Total size of the buffer
	writePtr uint32      // Next position to write (atomic)
	readPtr  uint32      // Next position to read (atomic)
}

// NewCircularQueue initializes a new circular queue
func NewCircularQueue(size int) *CircularQueue {
	return &CircularQueue{
		buffer:   make([]byte, size),
		capacity: size,
		writePtr: 0,
		readPtr:  0,
	}
}

// Write adds data to the queue (non-blocking, spins if full)
func (q *CircularQueue) Write(data []byte) bool {
	// Record format: [4-byte length | data]
	record := make([]byte, 4+len(data))
	record[0] = byte(len(data) >> 24)
	record[1] = byte(len(data) >> 16)
	record[2] = byte(len(data) >> 8)
	record[3] = byte(len(data))
	copy(record[4:], data)

	writePtr := atomic.LoadUint32(&q.writePtr)
	readPtr := atomic.LoadUint32(&q.readPtr)
	used := (writePtr - readPtr + uint32(q.capacity)) % uint32(q.capacity)
	available := uint32(q.capacity) - used - 1 // -1 to avoid writePtr == readPtr ambiguity

	if int(available) < len(record) {
		return false // Queue full, spin or retry
	}

	// Write the record
	for i := 0; i < len(record); i++ {
		q.buffer[int(writePtr)%q.capacity] = record[i]
		writePtr++
	}

	// Update write pointer atomically
	atomic.StoreUint32(&q.writePtr, writePtr)
	return true
}

// Read retrieves the next record from the queue (returns nil if empty)
func (q *CircularQueue) Read() []byte {
	readPtr := atomic.LoadUint32(&q.readPtr)
	writePtr := atomic.LoadUint32(&q.writePtr)

	if readPtr == writePtr {
		return nil // Queue empty
	}

	// Read the length (4 bytes)
	length := int(q.buffer[int(readPtr)%q.capacity])<<24 |
		int(q.buffer[(int(readPtr)+1)%q.capacity])<<16 |
		int(q.buffer[(int(readPtr)+2)%q.capacity])<<8 |
		int(q.buffer[(int(readPtr)+3)%q.capacity])
	readPtr += 4

	// Read the data
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = q.buffer[int(readPtr)%q.capacity]
		readPtr++
	}

	// Update read pointer atomically
	atomic.StoreUint32(&q.readPtr, readPtr)
	return data
}

// Simulated network receiver (producer)
func a(queue *CircularQueue) {
	for i := 0; ; i++ {
		data := []byte(fmt.Sprintf("Network data packet #%d", i))
		for !queue.Write(data) {
			time.Sleep(1 * time.Microsecond) // Spin briefly if queue is full
		}
		fmt.Printf("a() wrote: %s\n", data)
		time.Sleep(100 * time.Millisecond) // Simulate network delay
	}
}

// Simulated processor (consumer)
func b(queue *CircularQueue) {
	for {
		data := queue.Read()
		if data == nil {
			time.Sleep(1 * time.Microsecond) // Spin briefly if queue is empty
			continue
		}
		fmt.Printf("b() processed: %s\n", data)
		time.Sleep(150 * time.Millisecond) // Simulate processing delay
	}
}

func main() {
	queue := NewCircularQueue(1024) // 1KB buffer

	// Start producer and consumer as goroutines
	go a(queue)
	go b(queue)

	// Keep the program running
	select {}
}
