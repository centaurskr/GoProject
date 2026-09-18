//
// Proxy monitor function
// 
package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/logout"
)

// go routine
func RouterMonitor(monsock *zmq.Socket){
	for {
		event, _, _, err := monsock.RecvEvent(0)
		if err != nil{
			Log.Error("Router Monitor : %v", err)
			continue
		}
		switch event {
			case zmq.EVENT_ACCEPTED :
				Log.Infor("\tRouter Accepted ")
			case zmq.EVENT_DISCONNECTED :
				Log.Infor("\tRouter Disconnected ")
		}
	}
}

func DealerMonitor(monsock *zmq.Socket){
	for {
		event, _, _, err := monsock.RecvEvent(0)
		if err != nil{
			Log.Error("Dealer Monitor : %v", err)
			continue
		}
		switch event {
			case zmq.EVENT_ACCEPTED :
				Log.Infor("\tDealer Accepted ")
			case zmq.EVENT_DISCONNECTED :
				Log.Infor("\tDealer Disconnected ")
		}
	}
}
