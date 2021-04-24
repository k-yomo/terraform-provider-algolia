default: testacc

.PHONY: generate
generate:
	go generate ./...

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -coverprofile=coverage.out -timeout 120m
	go tool cover -html=coverage.out
	go tool cover -func=coverage.out

.PHONY: fmt
fmt:
	go fmt ./...
	terraform fmt --recursive