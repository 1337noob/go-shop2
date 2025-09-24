migrate:
	migrate -path schema -database "postgresql://order_history:order_history@127.0.0.1:5439/order_history?sslmode=disable" -verbose up

rollback:
	migrate -path schema -database "postgresql://order_history:order_history@127.0.0.1:5439/order_history?sslmode=disable" -verbose down

make-migration:
	migrate create -ext sql -dir schema -seq $(name)

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto
.PHONY: migrate rollback make-migration proto
