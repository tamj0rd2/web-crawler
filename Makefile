test: build
	go test ./... --race

build:
	go build -o crawler main.go

manual-test: build
	./crawler https://monzo.com > stdout.jsonl
