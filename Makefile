coverage:
	go test ./...  -coverpkg=./... -coverprofile ./coverage.out
	go tool cover -func ./coverage.out

run-example:
	go run example/multistep-example/main.go 
	go run example/dynamic-steps-example/main.go 