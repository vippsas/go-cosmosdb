example:
	go build -o ./dist/bin/cosmosapi-examples ./examples/cosmosapi/main.go
	go build -o ./dist/bin/cosmos-examples ./examples/cosmos/main.go

test:
	go build cmd/cosmosdb-apply/main.go
	go test -v `go list ./cosmosapi`
	go test -v `go list ./cosmos`
	go test -v `go list ./cosmostest`
