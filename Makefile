cover: cover-profile cover-html

check-swagger:
	which swagger || go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger: check-swagger
	$$(go env GOPATH)/bin/swagger generate spec --exclude-deps -o ./docs/swagger.yaml --scan-models

serve-swagger: swagger
	$$(go env GOPATH)/bin/swagger serve -F=swagger ./docs/swagger.yaml

test:
	@go test -v ./...

cover-profile:
	@go test -v -coverprofile cover.out ./...

cover-html:
	@go tool cover -html=cover.out -o cover.html