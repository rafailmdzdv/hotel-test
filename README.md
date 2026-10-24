# записная книжка отеля

## запуск

### docker

```bash
docker compose up --build
```

работает на http://localhost:8080

```bash
# остановка
docker compose down
```

---

### локально

```bash
cp .env.example .env
go run ./cmd/server
```
---

## тесты

```bash
go test -v -race ./...
```

## endpoints

#### 1. healthcheck

```bash
curl -X GET http://localhost:8080/health
```

#### 2. регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "hotel_admin",
    "password": "strongPassword123"
  }'
```

#### 3. получение токена

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "hotel_admin",
    "password": "strongPassword123"
  }'
```

#### 4. создание записи

```bash
curl -X POST http://localhost:8080/api/v1/guests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Иванов Иван Иванович",
    "phone": "+7 (999) 123-45-67",
    "room_number": "402",
    "check_in_date": "2026-11-01",
    "check_out_date": "2026-11-07",
    "note": "абв"
  }'
```

#### 5. получение списка гостей

```bash
# получить первые 20 записей
curl -X GET "http://localhost:8080/api/v1/guests?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# фильтрация по номеру комнаты
curl -X GET "http://localhost:8080/api/v1/guests?room_number=402" \
  -H "Authorization: Bearer $TOKEN"

# фильтрация по части номера телефона
curl -X GET "http://localhost:8080/api/v1/guests?phone=123-45-67" \
  -H "Authorization: Bearer $TOKEN"
```

#### 6. получение записи по id

```bash
curl -X GET http://localhost:8080/api/v1/guests/18c8e964-b040-424a-bb13-f4a7c152a514 \
  -H "Authorization: Bearer $TOKEN"
```

#### 7. обновление записи

```bash
curl -X PUT http://localhost:8080/api/v1/guests/18c8e964-b040-424a-bb13-f4a7c152a514 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Иванов Иван Иванович",
    "phone": "+7 (999) 123-45-67",
    "room_number": "402",
    "check_in_date": "2026-11-01",
    "check_out_date": "2026-11-10",
    "note": "Продлил проживание до 10 ноября"
  }'
```

#### 8. удаление записи

```bash
curl -X DELETE http://localhost:8080/api/v1/guests/18c8e964-b040-424a-bb13-f4a7c152a514 \
  -H "Authorization: Bearer $TOKEN"
```

## front

расположен в submodule `hotel-test-front`
