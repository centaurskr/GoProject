all : goTimeRandom
SRC = timeRandom.go
goTimeRandom: $(SRC)
	go build -o goTimeRandom $(SRC)
clean :
	rm -f goTimeRandom
	touch $(SRC)

install :
	@make clean
	make goTimeRandom
	strip goTimeRandom
	mv goTimeRandom  ~/bin
