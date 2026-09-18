all : goHogaRcv
goHogaRcv : hoga.go
	go build -o goHogaRcv hoga.go
clean :
	rm -f goHogaRcv
	touch hoga.go

install :
	@make clean
	make goHogaRcv
	strip goHogaRcv
	mv goHogaRcv ~/bin

