example:
	go build -o ./dist/bin/examples ./examples/main.go

test:
	go test -v `go list ./`
