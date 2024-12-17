all:
	$(error please pick a target)

lint:
	staticcheck ./...
	gosec ./...
	govulncheck ./...

test: lint
	go test -v

gosec_url := https://github.com/securego/gosec/releases/download/v2.21.4/gosec_2.21.4_linux_amd64.tar.gz
install-tools:
	curl -L $(gosec_url) | tar xz -C $(shell go env GOPATH)/bin gosec
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/caarlos0/svu@latest
	@echo "Done. Don't forget to add '\$$(go env GOPATH)/bin' to your '\$$PATH'"

ci: install-tools test

release-patch:
	git tag $(shell svu patch)
	git push --tags

release-minor:
	git tag $(shell svu minor)
	git push --tags

build:
	goreleaser build --snapshot --clean
