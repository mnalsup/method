install: deps
	go install ./...
deps:
	go mod download
