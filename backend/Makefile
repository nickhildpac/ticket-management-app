postgres:
	docker run --name ticketmgmt -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -d postgres
createdb:
	docker exec -it ticketmgmt createdb --username=postgres --owner=postgres ticket_management
dropdb:
	docker exec -it ticketmgmt dropdb ticket_management
migrateup:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose up
migratedown:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose down
migrateup1:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose up 1
migratedown1:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/ticket_management?sslmode=disable" -verbose down 1
sqlc:
	sqlc generate
test:
	go test -v -cover ./...
server:
	go run cmd/main.go
.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server
