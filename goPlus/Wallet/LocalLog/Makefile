all : goWallet
SRC = wallet.go  dbfunc.go
goWallet : $(SRC)
	go build -o goWallet  $(SRC)
clean :
	rm -f goWallet
	touch $(SRC)

install :
	@make clean
	make goWallet
	strip goWallet
	mv goWallet ~/bin
