all : goPubZmq
goPubZmq : pubzmq.go
	go build -o goPubZmq pubzmq.go
clean :
	rm -f goPubZmq
	touch pubzmq.go

install :
	@make clean
	make goPubZmq
	strip goPubZmq
	mv goPubZmq ~/bin

