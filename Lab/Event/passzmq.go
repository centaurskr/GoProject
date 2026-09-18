package main

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lzmq
#define ZMQ_BUILD_DRAFT_API
#include <stdio.h>
#include <zmq.h>  // Include zmq.h to use zmq socket in C

void printZmqSocket(void *socket){
	printf("zmqsocket [%p]\n", socket);
}

*/
import "C"
import (
    "fmt"
    zmq "github.com/pebbe/zmq4"
    "unsafe"
	"runtime"
)

func main() {
    // Create an instance of DATA
    s, err := zmq.NewSocket(zmq.REP)
    if err != nil {
        fmt.Println("Error creating ZMQ socket:", err)
        return
    }
    defer s.Close()

	runtime.KeepAlive(s)
    // Pass the pointer to C function
    C.printZmqSocket(C.uintptr_t(uintptr(unsafe.Pointer(s))))  // Pass the zmq.Socket
}

