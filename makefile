default: install

install:
	go build -o=./cmd/boil ./cmd/boil
	go install ./cmd/boil