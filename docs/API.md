# HCM API Reference

Base URL (dev): `http://localhost:8080/api`

All routes except `/login` require a JWT:

```
Authorization: Bearer <token>
```

The JWT carries `tenant_id`, so **never** send a tenant id in the request body — it is ignored.

---

## Auth

### `POST /login`

```json
// request
{ "email": "admin@acme.com", "password": "Password@123" }
```

```json
// 200 OK
{
  "token": "eyJhbGciOi...",
  "user": {
    "id": "0a8c…",
    "email": "admin@acme.com",
    "name": "Acme Corp Admin",
    "role": "ADMIN",
    "tenant_id": "b1f1…"
  }
}
```

Errors: `401 invalid credentials`, `400 validation error`.

### `GET /me`

Returns the authenticated user (same shape as `user` above). Useful for re-hydrating the frontend on reload.

---

## Employee Field Configuration

Stored per tenant.

### `GET /employee-fields[?active_only=true]`

Roles: ADMIN, HR.

```json
[
  {
    "id": "…",
    "tenant_id": "…",
    "field_name": "Work Mode",
    "field_key": "work_mode",
    "field_type": "dropdown",
    "required": true,
    "active": true,
    "options": ["Office", "Hybrid", "Remote"],
    "display_order": 1,
    "created_at": "…",
    "updated_at": "…"
  }
]
```

`field_type` ∈ `text | number | date | dropdown | boolean | email | phone`.

### `POST /employee-fields`  (ADMIN only)

```json
{
  "field_name": "Work Mode",
  "field_key": "work_mode",
  "field_type": "dropdown",
  "required": true,
  "active": true,
  "options": ["Office", "Hybrid", "Remote"],
  "display_order": 1
}
```

Rules:
- `field_key` must match `^[a-z][a-z0-9_]{0,99}$` and is unique per tenant.
- `field_type` must be one of the supported types.
- `dropdown` requires at least one option.

Errors: `400`, `409 field_key already exists`.

### `PUT /employee-fields/:id`  (ADMIN only)

Send only the fields you want to change. `field_key` is **immutable** to keep stored values consistent. To deactivate:

```json
{ "active": false }
```

---

## Employees

Roles: ADMIN, HR.

### `GET /employees`

Query params:

| Param        | Default | Description                                  |
|--------------|---------|----------------------------------------------|
| `page`       | 1       |                                              |
| `page_size`  | 20      | max 100                                      |
| `search`     | —       | matches name, email, employee_code (case-insensitive) |
| `department` | —       | exact match                                  |
| `status`     | —       | one of `ACTIVE | INACTIVE | ON_LEAVE | TERMINATED` |

Response:

```json
{
  "items": [ { /* EmployeeResponse */ } ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

### `GET /employees/:id`

Returns a single employee (or 404).

### `POST /employees`

```json
{
  "name": "Alice Walker",
  "email": "alice@acme.com",
  "phone": "+1 555 123 4567",
  "employee_code": "ACME-001",
  "department": "Engineering",
  "designation": "Senior Engineer",
  "date_of_joining": "2023-01-15T00:00:00Z",
  "employment_type": "Full-Time",
  "status": "ACTIVE",
  "custom_fields": {
    "work_mode": "Hybrid",
    "pan_number": "ABCDE1234F",
    "has_laptop": true
  }
}
```

Validation:
- `name`, `email`, `employee_code` required.
- `email` must be valid; uniqueness is per-tenant.
- `employee_code` uniqueness is per-tenant.
- `phone` (if present) must match `^\+?[0-9\s\-]{7,20}$`.
- `status` must be one of the allowed values (default `ACTIVE`).
- `custom_fields` is validated against the **active** field config:
  - required custom fields must be present and non-empty;
  - dropdowns must match an option;
  - numbers must parse to a number;
  - dates must be `YYYY-MM-DD` or RFC3339;
  - emails / phones use the same regexes as system fields;
  - **unknown keys are rejected**.

Errors: `400 validation`, `409 email or employee_code already exists for this tenant`.

### `PUT /employees/:id`

All fields optional. If `custom_fields` is sent, the values for currently-active fields are replaced; values for **inactive** fields are preserved.

---

## Response Shape: `EmployeeResponse`

```json
{
  "id": "…",
  "name": "Alice Walker",
  "email": "alice@acme.com",
  "phone": "+1 555 123 4567",
  "employee_code": "ACME-001",
  "department": "Engineering",
  "designation": "Senior Engineer",
  "date_of_joining": "2023-01-15T00:00:00Z",
  "employment_type": "Full-Time",
  "status": "ACTIVE",
  "custom_fields": { "work_mode": "Hybrid", "has_laptop": true },
  "created_at": "…",
  "updated_at": "…"
}
```

`custom_fields` values come back **typed**: numbers as numbers, booleans as booleans, the rest as strings.

---

## Error Format

```json
{ "error": "human readable message" }
```

Common status codes:

| Code | Meaning                                |
|------|----------------------------------------|
| 400  | Validation / bad input                  |
| 401  | Missing / invalid / expired token       |
| 403  | Authenticated but lacking the role      |
| 404  | Not found (or not in your tenant)       |
| 409  | Unique constraint violation             |
| 500  | Unexpected server error                 |

---

## cURL Examples

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@acme.com","password":"Password@123"}' | jq -r .token)

# List active fields
curl -H "Authorization: Bearer $TOKEN" \
  'http://localhost:8080/api/employee-fields?active_only=true'

# Create an employee
curl -X POST http://localhost:8080/api/employees \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"Bob",
    "email":"bob@acme.com",
    "employee_code":"ACME-002",
    "department":"Sales",
    "custom_fields": { "work_mode": "Office" }
  }'

# Search + filter + paginate
curl -H "Authorization: Bearer $TOKEN" \
  'http://localhost:8080/api/employees?search=ali&status=ACTIVE&page=1&page_size=10'
```
