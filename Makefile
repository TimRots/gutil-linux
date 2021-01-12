all:
	@env GOOS=linux GOARCH=amd64 go build -tags all -o bin/lsirq cmd/lsirq/lsirq.go
	@env GOOS=linux GOARCH=amd64 go build -tags all -o bin/lspci cmd/lspci/lspci.go
