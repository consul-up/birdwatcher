build-bins:
	mkdir -p bin && \
	cd backend && \
	GOOS=linux GOARCH=amd64 go build -o ../bin/backend-linux-amd64 ./... && \
	GOOS=linux GOARCH=arm64 go build -o ../bin/backend-linux-arm64 ./... && \
	GOOS=darwin GOARCH=amd64 go build -o ../bin/backend-darwin-amd64 ./... && \
	GOOS=darwin GOARCH=arm64 go build -o ../bin/backend-darwin-arm64 ./... && \
	cd ../frontend && \
	GOOS=linux GOARCH=amd64 go build -o ../bin/frontend-linux-amd64 ./... && \
	GOOS=linux GOARCH=arm64 go build -o ../bin/frontend-linux-arm64 ./... && \
	GOOS=darwin GOARCH=amd64 go build -o ../bin/frontend-darwin-amd64 ./... && \
	GOOS=darwin GOARCH=arm64 go build -o ../bin/frontend-darwin-arm64 ./...
