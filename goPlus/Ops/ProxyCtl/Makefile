goProxyCtl : proxyctl.go
	go build -o goProxyCtl proxyctl.go
clean :
	rm -f goProxyCtl
	touch proxyctl.go
install :
	make clean
	make  goProxyCtl
	strip goProxyCtl
	mv goProxyCtl ~/bin
