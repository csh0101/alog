export GOPROXY='goproxy.cn,direct'
export CGO_ENABLED=0
export GO111MODULE=on


default: lint test


lint:
	@echo "> Lint"
	golangci-lint run

test: 
	@echo "> Testing"
	go test ./...
