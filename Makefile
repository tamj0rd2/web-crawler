test: build
	go test ./...

build:
	go build -o crawler main.go

manual-test: build
	./crawler https://monzo.com