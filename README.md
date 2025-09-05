start docker containers
```shell
docker compose up -d
```

migrate order db
```shell
cd order
make migrate
```

migrate product db
```shell
cd product
make migrate
```

migrate inventory db
```shell
cd inventory
make migrate
```

migrate payment db
```shell
cd payment
make migrate
```

migrate order_saga db
```shell
cd order_saga
make migrate
```

start order service
```shell
go run order/cmd/main.go
```

start product service
```shell
go run product/cmd/main.go
```

start inventory service
```shell
go run inventory/cmd/main.go
```

start payment service
```shell
go run payment/cmd/main.go
```

start order_saga service
```shell
go run order_saga/cmd/main.go
```

http://localhost:8080/ui/clusters/local-kafka/all-topics?perPage=25


user: product  
password: product  
jdbc:postgresql://localhost:5433/product  

user: inventory  
password: inventory  
jdbc:postgresql://localhost:5434/inventory  

user: payment  
password: payment  
jdbc:postgresql://localhost:5435/payment  

user: order  
password: order  
jdbc:postgresql://localhost:5436/order  

user: saga  
password: saga  
jdbc:postgresql://localhost:5437/saga  

