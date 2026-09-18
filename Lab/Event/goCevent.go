package main
/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lzmq
#define ZMQ_BUILD_DRAFT_API
#include <stdio.h>
#include <stdlib.h>
#include <zmq.h>
void *Poller;

void InitPoller() {
    Poller = zmq_poller_new();
}

int AddPoller(void *data) {
    POLLDATA *polldata = (POLLDATA *)data;
    if (polldata->sockfd >= 0)
        return zmq_poller_add_fd(Poller, polldata->sockfd, data, ZMQ_POLLIN);
    else 
        return zmq_poller_add(Poller, polldata->zmqsock, data, ZMQ_POLLIN);
}
int DeletePoller(void *data){
    POLLDATA *polldata = (POLLDATA *)data;
    if (polldata->sockfd >= 0)
        return zmq_poller_remove_fd(Poller, polldata->sockfd);
    else 
        return zmq_poller_remove(Poller, polldata->zmqsock);
}
void *PollerWait() { 
    zmq_poller_event_t events[1];
    int n = zmq_poller_wait_all(Poller, events, 1, 1000);
    if (n <= 0) return NULL;
    return events[0].user_data;
}
*/
import "C"
import (
	//"net"
	"fmt"
	"unsafe"
	"syscall"
	zmq "github.com/pebbe/zmq4"
)
type POLLDATA struct {
    int   sockfd;
    void *zmqsock;
}

func sockHandler(fd int) {
    fmt.Println("Handling network socket event for fd:", fd)
}

func zmqHandler(sock *zmq.Socket) {
    fmt.Println("Handling ZMQ socket event")
}


func main() {
	var sockList map[uintptr]C.POLLDATA

    C.InitPoller()

    fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
    if err != nil {
        fmt.Println("Error creating socket:", err)
        return
    }
	sockList[uintptr(fd)] = C.POLLDATA{
		sockfd : C.int(fd),
		zmqsock : nil,
	}
	/*
    sockdata := &C.POLLDATA{
        sockfd:  C.int(fd),
        zmqsock: nil,
    }
	*/
    C.AddPoller(unsafe.Pointer(&sockList[uintptr(fd)]))

    sock, _ := zmq.NewSocket(zmq.SUB)
    sockList[uintptr(unsafe.Pointer(sock))] = C.POLLDATA{
        sockfd:  -1,
        zmqsock: unsafe.Pointer(sock),
    }
    C.AddPoller(unsafe.Pointer(&sockList[unsafe.Pointer(sock)]))

    for {
        ent := C.PollerWait()
        if ent == nil {
            continue
        }

        polldata := (*C.POLLDATA)(ent)
        if polldata.sockfd != -1 {
            sockHandler(int(polldata.sockfd))
        } else {
            zmqHandler((*zmq.Socket)(polldata.zmqsock))
        }
    }
}

