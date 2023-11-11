.PHONY=build_concurrency_get

build_concurrency_get:
	@go build -o bin/main concurrency_get/main.go

run_concurrency_get: build_concurrency_get
	@./bin/main

build_first_response:
	@go build -o bin/main first_response/main.go

run_first_response: build_first_response
	@./bin/main