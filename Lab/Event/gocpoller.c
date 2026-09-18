///
/// @brief  :
/// @file   : gocpoller.c
/// @date   : 2025. 02. 10. (월) 11:32:47 KST
/// @author : Cento
///
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lzmq
#define ZMQ_BUILD_DRAFT_API
#include <stdio.h>
#include <stdlib.h>
#include <zmq.h>
typedef struct {
    int   sockfd;
    void *zmqsock;
} POLLDATA;

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
void DeletePoller(void *data){
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
