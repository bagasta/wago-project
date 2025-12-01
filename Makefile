.PHONY: run build clean

run:
	cd backend && go run cmd/server/main.go

build:
	cd backend && go build -o bin/server cmd/server/main.go

clean:
	rm -rf backend/bin
