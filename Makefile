GOOS=linux
GOARCH=amd64
common = $(wildcard trillianlambda/*.go)

.PHONY : clean
clean:
	rm -f treesigner leafqueuer treesigner.zip leafqueuer.zip

treesigner.zip: $(common) trillianlambda/treesigner/signer.go
	go build -o treesigner trillianlambda/treesigner/signer.go
	zip treesigner.zip treesigner

leafqueuer.zip: $(common) trillianlambda/leafqueuer/handler.go
	go build -o leafqueuer trillianlambda/leafqueuer/handler.go
	zip leafqueuer.zip leafqueuer

all: treesigner.zip leafqueuer.zip

.DEFAULT_GOAL := all
