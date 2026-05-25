# HCM (Human Capital Management) — Multitenant Full-Stack App

A small, clean, production-style full-stack HCM application.

- **Backend** – Go (Gin + GORM) + PostgreSQL + JWT
- **Frontend** – React (Vite) + React Router + React Hook Form + Axios
- **Multitenant** – every table is scoped by `tenant_id`, derived from the JWT
- **Dynamic forms** – the employee form is rendered from per-tenant field config

---

## Quick Start (Docker — easiest)

Requires Docker Desktop.

```bash
# 1. Build & start postgres + backend + frontend
docker compose up -d --build

# 2. Seed demo tenants/users (run once)
docker compose run --rm seed

# 3. Open the app
#    http://localhost:5173
```

**Demo accounts (after seed):**

| Tenant  | Email              | Password       | Role  |
|---------|--------------------|----------------|-------|
| Acme    | `admin@acme.com`   | `Password@123` | ADMIN |
| Acme    | `hr@acme.com`      | `Password@123` | HR    |
| Globex  | `admin@globex.com` | `Password@123` | ADMIN |
| Globex  | `hr@globex.com`    | `Password@123` | HR    |

> Acme already has 3 sample custom fields and 1 sample employee. Globex starts blank.

---

## Quick Start (Local Dev)

### Prerequisites
- Go 1.22+
- Node 18+ / npm
- PostgreSQL 14+

### 1. Database

```sql
CREATE DATABASE hcm;
```

### 2. Backend

```bash
cd backend
cp .env.example .env       # edit DB_* values if needed
go mod tidy
go run ./cmd/seed          # one-time: seed demo data
go run ./cmd/server        # starts on :8080
```

### 3. Frontend

```bash
cd frontend
npm install
npm run dev                # starts on :5173 with proxy to :8080
```

Open `http://localhost:5173`, sign in with one of the seeded accounts.

---

## What's Included

### Authentication & Authorization
- `POST /api/login` returns a JWT containing `user_id`, `tenant_id`, `role`
- All other API routes require `Authorization: Bearer <token>`
- Two roles:
  - **ADMIN** – full access including field configuration
  - **HR** – everything except field configuration
- Passwords are bcrypt-hashed; hashes are never returned

### Multitenancy
- Every table carries `tenant_id`
- `tenant_id` always comes from the JWT — clients can never spoof it
- Unique indexes are `(tenant_id, …)` so two tenants can have the same email/code

### Dynamic Employee Form
- Backend exposes `GET /api/employee-fields?active_only=true`
- Frontend renders a form that is system-fields + tenant-fields
- Validation rules (required / email / phone / number / dropdown options / dates) are computed from the field config — the form is **not hardcoded**
- Server **re-validates** every payload (defense in depth)
- Unknown custom keys are rejected with a clear error
- Deactivating a field hides it from new forms but keeps existing data

### Employee Management
- Create / Read / Update / List
- List supports pagination (`page`, `page_size`), search (name/email/code), filter by `department`, filter by `status`

### Frontend
- Protected routes via React Router + Auth context
- Role-based menu (ADMIN sees Field Configuration link, HR does not)
- Loading / error states everywhere
- Clean, minimal CSS — no UI framework needed

---

## Project Structure

```
HCM Assignment/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go        # API entrypoint
│   │   └── seed/main.go          # Demo data seeder
│   ├── internal/
│   │   ├── auth/                 # JWT + password hashing
│   │   ├── config/               # env loader
│   │   ├── database/             # gorm + auto-migrate
│   │   ├── dto/                  # request/response types
│   │   ├── handlers/             # HTTP handlers
│   │   ├── middleware/           # JWTAuth, RequireRole
│   │   ├── models/               # GORM models
│   │   ├── routes/               # Gin route wiring
│   │   └── validator/            # custom-field validation
│   ├── migrations/001_init.sql   # reference schema
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/client.js         # axios + auth interceptor
│   │   ├── context/AuthContext.jsx
│   │   ├── components/
│   │   │   ├── DynamicField.jsx  # renders a single dynamic input
│   │   │   ├── Layout.jsx        # sidebar shell
│   │   │   └── ProtectedRoute.jsx
│   │   ├── pages/
│   │   │   ├── Login.jsx
│   │   │   ├── EmployeeList.jsx
│   │   │   ├── EmployeeForm.jsx  # dynamic create/edit
│   │   │   └── FieldConfig.jsx   # ADMIN-only
│   │   ├── App.jsx
│   │   ├── main.jsx
│   │   └── styles.css
│   ├── Dockerfile
│   └── package.json
├── docs/
│   └── API.md                    # full API reference
├── docker-compose.yml
└── README.md (this file)
```

---

## Database Tables

| Table                         | Purpose                                  |
|-------------------------------|------------------------------------------|
| `tenants`                     | One row per company                      |
| `users`                       | Login accounts; `role` = ADMIN \| HR     |
| `employee_custom_fields`      | Per-tenant field config                  |
| `employees`                   | Common employee columns                  |
| `employee_custom_field_values`| Custom values keyed by `(employee, field)` |

DDL reference: `backend/migrations/001_init.sql`. The Go server actually runs `gorm.AutoMigrate` on startup so the database is created automatically.

---

## API Reference

See [`docs/API.md`](docs/API.md) for the complete request/response specs.

## Database Reference

See [`docs/DATABASE.md`](docs/DATABASE.md) for:
- full schema/table definitions
- migration inventory used in this project
- ER diagram of table relationships

Quick summary:

| Method | Path                          | Role         |
|--------|-------------------------------|--------------|
| POST   | `/api/login`                  | public       |
| GET    | `/api/me`                     | any auth     |
| GET    | `/api/employee-fields`        | ADMIN, HR    |
| POST   | `/api/employee-fields`        | ADMIN        |
| PUT    | `/api/employee-fields/:id`    | ADMIN        |
| GET    | `/api/employees`              | ADMIN, HR    |
| POST   | `/api/employees`              | ADMIN, HR    |
| GET    | `/api/employees/:id`          | ADMIN, HR    |
| PUT    | `/api/employees/:id`          | ADMIN, HR    |

---

## How Tenant Isolation Works (in 3 sentences)

1. Login looks the user up by email, verifies the password, then bakes the user's `tenant_id` into the JWT.
2. Every authenticated handler reads `tenant_id` from `middleware.TenantID(c)` (which reads the parsed JWT claim) — never from the body or query string.
3. Every DB query filters on `WHERE tenant_id = ?`, and every unique index is composite on `(tenant_id, …)`, so two tenants are completely siloed.

---

## How Dynamic Forms Work (end to end)

1. Admin defines fields in **Field Configuration** (`employee_custom_fields`).
2. Frontend calls `GET /api/employee-fields?active_only=true` and renders one input per field via `DynamicField.jsx`.
3. Frontend sends `custom_fields: { key: value, … }` in the create/update payload.
4. Backend validates each value against the field's type and options. **Unknown keys are rejected.**
5. Each value is stored in `employee_custom_field_values` keyed by `(employee_id, field_id)`.

When a field is **deactivated**, it disappears from new forms but historic data is untouched.

---

## Notes & Trade-offs

- Custom values are stored as `TEXT` and re-typed on read. This keeps the schema simple and works for all listed types. For very large datasets you'd switch to typed columns or JSONB.
- No refresh-token flow yet — JWTs expire in 24h. Adding one is just two endpoints.
- The seed script is idempotent (skips existing rows).
- Frontend uses the Vite dev proxy in development, and Nginx as a reverse proxy in the Docker image.

---

## License

MIT — feel free to use as a reference.
