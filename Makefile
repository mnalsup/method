install: deps
	go install ./...
build: deps
	go build -o method ./cmd/method
deps:
	go mod download
