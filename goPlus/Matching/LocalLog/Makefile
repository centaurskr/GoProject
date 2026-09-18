
SOURCE = MatMain.go afterfunc.go fdsfunc.go memefunc.go \
	sisefunc.go clsmapfile.go 
goMatching : $(SOURCE)
	go build -o $@ $(SOURCE)
	strip $@

clean : 
	@rm  -f goMatching 
install:
	@make clean
	@make goMatching
	@mv goMatching ~/bin
