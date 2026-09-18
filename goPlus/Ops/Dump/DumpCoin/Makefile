TARGET=goDumpCoin
all : $(TARGET)

goDumpCoin : dumpcoin.go
	go build -o goDumpCoin dumpcoin.go
clean :
	rm -f $(TARGET)
	touch dumpcoin.go 
install :
	make clean
	make $(TARGET)
	strip $(TARGET)
	mv $(TARGET) ~/bin/.
