all : goNewCoin
goNewCoin : newcoin.go
	go build -o goNewCoin newcoin.go
clean :
	rm -f goNewCoin
	touch newcoin.go

install :
	@make clean
	make goNewCoin
	strip goNewCoin
	mv goNewCoin ~/bin

