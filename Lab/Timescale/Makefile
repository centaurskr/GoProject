all : goTimescale
SRC = timesc.go
goTimescale: $(SRC)
	go build -o goTimescale $(SRC)
clean :
	rm -f goTimescale
	touch $(SRC)

install :
	@make clean
	make goTimescale
	strip goTimescale
	mv goTimescale  ~/bin
