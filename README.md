# Практическое задание 3
## Шишков А.Д. ЭФМО-02-22
## Тема
Структурированное логирование в серверных приложениях.

## Цель
Научиться внедрять структурированные логи в сервис и применять единый стандарт логирования для диагностики и эксплуатации.

---

## 1. Выбор логгера

Выбран **zap** (`go.uber.org/zap`).

Причины:

- Выдаёт JSON по умолчанию — логи сразу готовы для парсинга (ELK, Loki, jq).
- Высокая производительность — zero-allocation в hot path, что важно для микросервисов под нагрузкой.
- Типизированные поля (`zap.String`, `zap.Int`, `zap.Error`) исключают случайную запись чувствительных данных через `fmt.Sprintf`.
- Широко используется в production Go-проектах.

---

## 2. Стандарт полей логов

Во всех сервисах (auth, tasks) используется единая схема полей:

### Обязательные поля

| Поле | Тип | Описание |
|------|-----|----------|
| `level` | string | Уровень: debug / info / warn / error |
| `ts` | string | Время события (ISO 8601) |
| `msg` | string | Описание события |
| `service` | string | Имя сервиса: `auth` или `tasks` |
| `request_id` | string | X-Request-ID (из заголовка или сгенерированный UUID) |

### Поля access log (middleware)

| Поле | Тип | Описание |
|------|-----|----------|
| `method` | string | HTTP-метод (GET, POST, PATCH, DELETE) |
| `path` | string | Путь запроса (`/v1/tasks`) |
| `status` | int | Код ответа (200, 401, 503...) |
| `duration_ms` | float | Длительность обработки в миллисекундах |
| `remote_ip` | string | IP-адрес клиента |

### Поля для ошибок и контекста

| Поле | Тип | Описание |
|------|-----|----------|
| `error` | string | Текст ошибки (без секретов) |
| `component` | string | Компонент: `handler`, `auth_middleware`, `auth_client_http`, `auth_client_grpc` |
| `task_id` | string | ID задачи (при CRUD-операциях) |
| `subject` | string | Субъект из токена (при успешной верификации) |
| `has_auth` | bool | Факт наличия токена (без самого значения) |

### Что запрещено логировать

- Пароли
- Токены доступа и refresh-токены
- Секреты и ключи
- Содержимое cookies

---

## 3. Примеры лог-событий

### 3.1. Успешный запрос — создание задачи


```bash
curl -i -X POST http://<SERVER_IP>:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: pz19-002" \
  -d '{"title":"Logs","description":"Implement zap","due_date":"2026-01-12"}'
```

201 Created. В логах Tasks видно цепочку: calling auth HTTP verify → auth HTTP verify: success → token verified → task created → request completed. В логах Auth — token verified → request completed (200).

<img width="1898" height="126" alt="image" src="https://github.com/user-attachments/assets/4c229940-2ec2-48ad-99dc-197300437c88" /> 


### 3.2. Запрос с ошибкой — невалидный токен

```bash
curl -i http://<SERVER_IP>:8082/v1/tasks \
  -H "Authorization: Bearer invalid-token" \
  -H "X-Request-ID: pz19-003"
```

401 Unauthorized. В логах Tasks уровень `warn` — `auth HTTP verify: unauthorized`, `invalid token`. Токен не попадает в логи (только `has_auth: true`). В логах Auth — `token verification failed` (уровень `warn`).


<img width="1904" height="300" alt="image" src="https://github.com/user-attachments/assets/3de613f4-9731-4856-8521-5ff4e3939d73" /> 


### 3.3. Межсервисный вызов — корреляция по request-id

Шаг 1 — получить токен:

```bash
curl -s -X POST http://<SERVER_IP>:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: pz19-010" \
  -d '{"username":"student","password":"student"}'
```

Шаг 2 — создать задачу (request-id видно в обоих сервисах):

```bash
curl -i -X POST http://<SERVER_IP>:8082/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer demo-token" \
  -H "X-Request-ID: pz19-011" \
  -d '{"title":"Cross-service","description":"Test correlation","due_date":"2026-01-12"}'
```

По `request_id: "pz19-011"` можно найти связанные события в логах обоих сервисов: Auth (`service: auth`, verify) и Tasks (`service: tasks`, task created).

<img width="1899" height="407" alt="image" src="https://github.com/user-attachments/assets/02aebe24-4ba9-4a27-bdba-03332d343e5c" /> 


---

## 4. Инструкция запуска и проверки

### Установка

```bash
git clone https://github.com/Alex171228/pz3.2.git
cd pz3.2
go mod download
```

### Запуск (2 терминала)

**Терминал 1 — Auth Service:**

```bash
export AUTH_PORT=8081
export AUTH_GRPC_PORT=50051
go run ./services/auth/cmd/auth
```

**Терминал 2 — Tasks Service:**

```bash
export TASKS_PORT=8082
go run ./services/tasks/cmd/tasks
```

### Проверка (3-й терминал)

Выполните запросы из пункта 3 и посмотрите логи в консоли сервера для проверки работы.

---

## 5. Контрольные вопросы

**1. Почему структурированные логи удобнее строковых?**

Структурированные логи (JSON) можно парсить, фильтровать и агрегировать автоматически. Строковые логи требуют regex для извлечения данных, что ненадёжно и медленно.

**2. Что такое request-id и как он помогает при диагностике?**

X-Request-ID — уникальный идентификатор запроса, который прокидывается через все сервисы в цепочке вызовов. Позволяет по одному ID найти все связанные лог-события в разных сервисах.

**3. Какие поля вы считаете обязательными для access log?**

`request_id`, `method`, `path`, `status`, `duration_ms`, `service`. Без них невозможно ответить на базовые вопросы: какой запрос, куда, как быстро, с каким результатом.

**4. Почему нельзя писать токены и пароли в логи?**

Логи часто хранятся в открытом виде, реплицируются в системы мониторинга, доступны широкому кругу разработчиков. Утечка токена из логов = компрометация пользователя.

**5. Что логировать в ERROR, а что в INFO/WARN?**

- **INFO** — штатные события: запрос завершён, задача создана, сервер запущен.
- **WARN** — ожидаемые проблемы: невалидный токен, задача не найдена, неверный формат.
- **ERROR** — неожиданные сбои: auth недоступен, ошибка БД, таймаут межсервисного вызова.
