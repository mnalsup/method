install: deps
	go install ./cmd/method
build: deps
	go build -o method ./cmd/method
deps:
	go mod download
