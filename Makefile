build:
	@go build -o bin/RFID-mikrotik-db cmd/main.go

run: build
	@./bin/RFID-mikrotik-db

test:
	@go test -v ./...


migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down