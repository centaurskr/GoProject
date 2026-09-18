goPubSubProxy : PubSubProxy.go
	go build -o goPubSubProxy PubSubProxy.go
clean :
	@rm -f goPubSubProxy
	@touch PubSubProxy.go
install : 
	@make clean
	@make 
	@strip goPubSubProxy
	@mv goPubSubProxy ~/bin/
