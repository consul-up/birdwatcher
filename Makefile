build-bins:
	mkdir -p bin && \
	cd backend && \
	GOOS=linux GOARCH=amd64 go build -o ../bin/backend ./... && \
	cd ../frontend && \
	GOOS=linux GOARCH=amd64 go build -o ../bin/frontend ./...
