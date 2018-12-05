FROM golang:1.11.2-stretch as builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /cosmosdb-apply cmd/cosmosdb-apply/main.go

FROM scratch
COPY --from=builder cosmosdb-apply /
ENTRYPOINT ["/cosmosdb-apply"]