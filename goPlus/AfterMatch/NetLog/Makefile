all : goAfterMatch
SRC = afterMatch.go dataproc.go dbfunc.go jsonfunc.go spscqueue.go
goAfterMatch : $(SRC)
	go build -o goAfterMatch  $(SRC)
clean :
	rm -f goAfterMatch
	touch $(SRC)

install :
	@make clean
	make goAfterMatch
	strip goAfterMatch
	mv goAfterMatch ~/bin
