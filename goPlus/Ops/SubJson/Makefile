all : goSubJson
goSubJson : subjson.go
	go build -o goSubJson subjson.go
clean :
	rm -f goSubJson
	touch subjson.go

install :
	@make clean
	make goSubJson
	strip goSubJson
	mv goSubJson ~/bin

