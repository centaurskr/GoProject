// REQ/REP Proxy

package main
import (
	zmq "github.com/pebbe/zmq4"
	Log "centolib/logout"
	"centolib/tbzmq"
	"os"
	"time"
)

func main(){
	Log.InitLogger("", Log.ALL_MODE);defer Log.Close()	
	
	if len(os.Args) != 3{
		Log.Error("REQ/REP Proxy");
		Log.Error("Using : %s <front> <end>", os.Args[0])
		os.Exit(0)
	}

	router, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil{
		Log.Error("NewSocket (ROUTER)\n\t%v", err)
		os.Exit(0)
	}
	dealer, err := zmq.NewSocket(zmq.DEALER)
	if err != nil{
		Log.Error("NewSocket (DEALER)\n\t%v", err)
		os.Exit(0)
	}

	err = router.Bind(os.Args[1])
	if err != nil{
		Log.Error("router(REQ).Bind()-%s\n\t%v", os.Args[1], err)
		os.Exit(0)
	}
	err = dealer.Bind(os.Args[2])
	if err != nil{
		Log.Error("delaer(REP).Bind()-%s\n\t%v", os.Args[2], err)
		os.Exit(0)
	}
	time.Sleep(3000 * time.Microsecond)

	// Monitor socket for router
	rmonsock, err := tbzmq.MakeMonitorSocket("", router)
	if err != nil{
		Log.Error("MakeMonitorSocket(ROUTER)\n\t%v", err)
		os.Exit(0)
	}
	go RouterMonitor(rmonsock)

	// Monitor socket for dealer
	dmonsock, err := tbzmq.MakeMonitorSocket("", dealer)
	if err != nil{
		Log.Error("MakeMonitorSocket(DEALER)\n\t%v", err)
		os.Exit(0)
	}
	go DealerMonitor(dmonsock)

	Log.Infor("REQ/REP Proxy Server (Front:%s Back:%s)", os.Args[1], os.Args[2])
	zmq.Proxy(router, dealer, nil)

}
