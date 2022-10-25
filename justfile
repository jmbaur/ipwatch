run:
	go run ./cmd/ipwatch -debug -4 -hook=internal:echo -filter=!IsLoopback

build:
	go build -o $out/ipwatch ./cmd/ipwatch

check:
	go test ./...
	staticcheck ./...
