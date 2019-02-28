example:
	go build -o ./dist/bin/cosmosapi-examples ./examples/cosmosapi/main.go
	go build -o ./dist/bin/cosmos-examples ./examples/cosmos/main.go

test:
	go build cmd/cosmosdb-apply/main.go
	go test -v `go list ./cosmosapi`
	go test -tags=offline -v `go list ./cosmos`
	go test -v `go list ./cosmostest`

vet:
	go vet ./...
	GO111MODULE=off go get -d golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	GO111MODULE=off go build -o shadow golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
	go vet -vettool=shadow ./...

.PHONY: test vet
