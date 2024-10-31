coverage:
	go test $(go list ./... | grep -v /example/)  -coverpkg=./... -coverprofile ./coverage.out
	go tool cover -func ./coverage.out
