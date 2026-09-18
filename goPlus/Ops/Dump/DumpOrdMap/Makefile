goDumpOrdmap : dumpordmap.go
	go build -o goDumpOrdmap dumpordmap.go
clean :
	rm -f goDumpOrdmap
	touch dumpordmap.go 
install :
	make clean
	make goDumpOrdmap
	strip goDumpOrdmap
	mv goDumpOrdmap ~/bin/.
