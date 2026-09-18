TARGET = goDumpCond
goDumpCond : dumpcond.go
	go build -o goDumpCond dumpcond.go
clean :
	rm -f $(TARGET)
	touch dumpcond.go
install :
	make clean
	make $(TARGET)
	strip $(TARGET)
	mv $(TARGET) ~/bin/.
