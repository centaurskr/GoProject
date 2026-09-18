// lock-free SPSC(Single Producer Single Consumer) CircularQueue 구현.
package main

import (
	"sync/atomic"
)

// Qdata는 CircularQueue에 저장되는 단일 항목으로 타입 문자열과 바이트 데이터를 담는다.
type Qdata struct{
	tr    string
	data  []byte
}

// CircularQueue는 단일 생산자-단일 소비자 구조의 lock-free 원형 큐다.
type CircularQueue struct {
	data      []Qdata
	qSize     int         // Total size of the buffer
	writePos  int32       // Next position to write
	readPos   int32       // Next position to read
	readCh    chan struct{}
}

// NewCircularQueue는 지정한 size 크기의 CircularQueue를 생성하여 반환한다.
func NewCircularQueue(size int) *CircularQueue {
	return &CircularQueue{
		data:   make([]Qdata, size),
		qSize: size,
		writePos: 0,
		readPos:  0,
		readCh : make(chan struct{}, 1),
	}
}

// Write는 큐에 tr(타입)과 data를 기록한다. 큐가 가득 차면 false를 반환한다.
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

	q.readCh <- struct {}{}

	return true
}

// Read는 큐에서 쌓인 모든 항목을 한 번에 읽어 반환한다. 데이터가 없으면 nil, false를 반환한다.
func (q *CircularQueue)Read()([]Qdata, bool){
	wPos := atomic.LoadInt32(&q.writePos)
	rPos := atomic.LoadInt32(&q.readPos)
	if rPos == wPos{
		return  nil, false
	}
	var rData []Qdata

	for rPos != wPos {
		rData = append(rData, q.data[rPos])
		q.data[rPos].data = nil  // Hey GC! Clean memeory
		rPos = (rPos +1 ) % int32(q.qSize)
	}
	atomic.StoreInt32(&q.readPos, rPos)

	return rData, true
}

