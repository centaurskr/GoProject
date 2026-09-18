all : goAfter
SRC = after.go  queue4.go
#SRC = after.go  queue3.go
#SRC=after.go
goAfter: $(SRC)
	go build -o goAfter $(SRC)
clean :
	rm -f goAfter
	touch $(SRC)

install :
	@make clean
	make goAfter
	strip goAfter
	mv goAfter  ~/bin
