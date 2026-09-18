goExrate60 : exrate60.go
	go build -o goExrate60 exrate60.go
clean :
	rm -f goExrate60
	touch exrate60.go
install :
	make clean
	make goExrate60
	strip goExrate60
	mv goExrate60 ~/bin
