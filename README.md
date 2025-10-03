## Краткое описание сервисов:

- **API Gateway**

  Единая точка входа для всех запросов. Маршрутизирует запрос в соответствующий сервис.

- **Product**

  Управление товарами

- **Order**

  Управление заказами

- **Inventory**

  Управление остатками товаров на складе

- **Payment**

  Обработка платежей

- **Order Saga**

  Координирует все шаги создания заказа, обеспечивает согласованность данных

- **Order History**

  Хранит историю заказов. Read модель в CQRS

- **Broker**

  Kafka. Асинхронное взаимодействие между сервисами

![Схема](/docs/images/shop.drawio.png)

### Успешное создание заказа
1. Клиент -> API Gateway: POST /api/orders
2. API Gateway -> Broker: отправляет команду SagaCreateOrder 
3. Order Saga читает SagaCreateOrder: создает и запускает сагу, отправляет команду CreateOrder
4. Order читает CreateOrder: создает заказ и отправляет событие OrderCreated
5. Order Saga читает OrderCreated: сохраняет id заказа, отправляет команду ValidateProducts
6. Product читает ValidateProducts: получает названия и цены, отправляет событие ProductsValidated
7. Order Saga читает ProductsValidated: сохраняет названия и цены, отправляет команду ReserveInventory
8. Inventory читает ReserveInventory: уменьшает количество товаров на складе, отправляет событие InventoryReserved
9. Order Saga читает InventoryReserved: отправляет команду ProcessPayment
10. Payment читает ProcessPayment: обрабатывает платеж, отправляет событие PaymentCompleted
11. Order Saga читает PaymentCompleted: сохраняет данные платежа, отправляет команду CompleteOrder
12. Order читает CompleteOrder: изменяет статус, отправляет событие OrderCompleted
13. Order Saga читает OrderCompleted: завершает сагу
14. Order History читает OrderCreated, ProductsValidated, InventoryReserved, PaymentCompleted, OrderCompleted: обновляет данные заказа

### Реализованные паттерны
- **Saga** оркестратор управляет транзакциями и обеспечивает согласованность данных
- **Outbox** гарантирует отправку сообщения в брокер
- **Database per service** независимые бд у каждого сервиса, слабая связанность сервисов
- **API Gateway** единая точка входа для всех запросов с аутентификацией и авторизацией
- **CQRS** модель чтения Order History для получения заказов с пагинацией и фильтрами
- **Idempotent Consumer** игнорирует дублирующиеся сообщения из брокера

### Примеры запросов в коллекции postman:
shop2.postman_collection.json


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

migrate order_history db
```shell
cd order_history
make migrate
```

migrate gateway db
```shell
cd gateway
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

start order_history service
```shell
go run order_history/cmd/main.go
```

start gateway service
```shell
go run gateway/cmd/main.go
```

топики:
```
order-saga-commands
order-commands
order-events
product-commands
product-events
inventory-commands
inventory-events
payment-commands
payment-events
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

user: gateway  
password: gateway  
jdbc:postgresql://localhost:5438/gateway 

user: order_history  
password: order_history  
jdbc:postgresql://localhost:5439/order_history 
