goReqRepProxy : ReqRepProxy.go
	go build -o goReqRepProxy ReqRepProxy.go
clean :
	@rm -f goReqRepProxy
	@touch ReqRepProxy.go
install : 
	@make clean
	@make 
	@strip goReqRepProxy
	@mv goReqRepProxy ~/bin/
