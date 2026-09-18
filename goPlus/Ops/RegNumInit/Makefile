all : goReadRegnum goRegNumInit

goReadRegnum : ReadRegnum.go
	go build -o goReadRegnum ReadRegnum.go

goRegNumInit : RegNumInit.go
	go build -o goRegNumInit RegNumInit.go

clean :
	rm -f goReadRegnum goRegNumInit
	touch *.go

install :
	@make clean
	@make all
	mv goReadRegnum ~/bin
	mv goRegNumInit ~/bin
