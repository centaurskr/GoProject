all : goMatchingCtl
SRC_LIST=matchingCtl.go coinlist.go
goMatchingCtl : $(SRC_LIST)
	go build -o goMatchingCtl $(SRC_LIST)
clean :
	rm -f goMatchingCtl
	touch $(SRC_LIST)

install :
	@make clean
	make goMatchingCtl
	strip goMatchingCtl
	mv goMatchingCtl ~/bin
