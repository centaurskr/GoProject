GO=go
BINDIR = ~/bin
TARGET=goapisvr
FILES=goapisvr.go 

$(TARGET) : $(FILES)
	go build -o $(TARGET) $(FILES)

clean :
	@rm -f $(TARGET)
	touch $(FILES)

install :
	@make clean
	make $(TARGET)
	strip $(TARGET)
	mv $(TARGET) $(BINDIR)
