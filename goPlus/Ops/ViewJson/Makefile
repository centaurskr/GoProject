all : goViewJson
goViewJson : viewJson.go
	go build -o goViewJson viewJson.go
clean :
	rm -f goViewJson
	touch viewJson.go

install :
	@make clean
	make goViewJson
	strip goViewJson
	mv goViewJson ~/bin

