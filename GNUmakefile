default: testacc

VERSION=0.0.1-snapshot

# Build provider binary and place it plugins directory to be able to sideload the built provider.
.PHONY: sideload
sideload:
	go build -o ~/.terraform.d/plugins/registry.terraform.io/k-yomo/algolia/${VERSION}/darwin_amd64/terraform-provider-algolia

.PHONY: generate
generate:
	go generate ./...

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
