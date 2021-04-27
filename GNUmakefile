default: testacc

.PHONY: generate
generate:
	go generate ./...

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -coverprofile=coverage.out -timeout 120m

# Run acceptance tests and show coverages
.PHONY: testacc-cover
testacc-cover: testacc
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out

.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt --recursive