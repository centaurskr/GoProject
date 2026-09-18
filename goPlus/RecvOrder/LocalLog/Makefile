SRC=recvOrder.go dbfunc.go jsonfunc.go
all : goRecvOrder
goRecvOrder : $(SRC)
	go build -o goRecvOrder $(SRC)
clean :
	rm -f goRecvOrder
	touch $(SRC)

install :
	@make clean
	make goRecvOrder
	strip goRecvOrder
	mv goRecvOrder ~/bin
