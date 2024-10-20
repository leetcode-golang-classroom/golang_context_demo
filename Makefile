.PHONY=build_concurrency_get

build_concurrency_get:
	@go build -o bin/main concurrency_get/main.go

run_concurrency_get: build_concurrency_get
	@./bin/main

build_first_response:
	@go build -o bin/main first_response/main.go

run_first_response: build_first_response
	@./bin/main

build_done_channel:
	@go build -o bin/main done_channel/main.go

run_done_channel: build_done_channel
	@./bin/main