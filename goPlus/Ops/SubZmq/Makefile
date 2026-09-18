all : goSubZmq
goSubZmq : subzmq.go
	go build -o goSubZmq subzmq.go
clean :
	rm -f goSubZmq
	touch subzmq.go

install :
	@make clean
	make goSubZmq
	strip goSubZmq
	mv goSubZmq ~/bin

