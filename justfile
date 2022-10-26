build:
	go build -o $out/ipwatch ./cmd/ipwatch

check: build
	staticcheck ./...
	go test ./...

run:
	go run ./cmd/ipwatch -debug -4 -hook=internal:echo -filter=!IsLoopback
