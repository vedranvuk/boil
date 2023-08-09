default: install

install:
	go build -o=./cmd/boil ./cmd/boil
	go install ./cmd/boil

installprod:
	go build -o=./cmd/boil -ldflags "-s -w" ./cmd/boil
	go install ./cmd/boil
