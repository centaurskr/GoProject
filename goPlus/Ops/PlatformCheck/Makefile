all : goPlatformCheck
SRC=platformCheck.go proxyMenu.go ordCond.go coinord.go
goPlatformCheck : $(SRC)
	go build -o goPlatformCheck $(SRC)
clean :
	rm -f goPlatformCheck
	touch $(SRC)

install :
	@make clean
	make goPlatformCheck
	strip goPlatformCheck
	mv goPlatformCheck ~/bin

