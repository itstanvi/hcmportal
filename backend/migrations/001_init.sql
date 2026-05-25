-- Reference schema (the Go server runs GORM AutoMigrate; this file is for documentation
-- and for environments where you want to apply DDL manually).

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    name            VARCHAR(200) NOT NULL,
    role            VARCHAR(20) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users (tenant_id);

CREATE TABLE IF NOT EXISTS employee_custom_fields (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    field_name      VARCHAR(200) NOT NULL,
    field_key       VARCHAR(100) NOT NULL,
    field_type      VARCHAR(20)  NOT NULL,
    required        BOOLEAN NOT NULL DEFAULT FALSE,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    options         JSONB,
    display_order   INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, field_key)
);
CREATE INDEX IF NOT EXISTS idx_fields_tenant ON employee_custom_fields (tenant_id);

CREATE TABLE IF NOT EXISTS employees (
    id                  UUID PRIMARY KEY,
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    name                VARCHAR(200) NOT NULL,
    email               VARCHAR(255) NOT NULL,
    phone               VARCHAR(30),
    employee_code       VARCHAR(50) NOT NULL,
    department          VARCHAR(100),
    designation         VARCHAR(100),
    date_of_joining     TIMESTAMPTZ,
    employment_type     VARCHAR(50),
    status              VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email),
    UNIQUE (tenant_id, employee_code)
);
CREATE INDEX IF NOT EXISTS idx_emp_tenant ON employees (tenant_id);

CREATE TABLE IF NOT EXISTS employee_custom_field_values (
    id              UUID PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    field_id        UUID NOT NULL REFERENCES employee_custom_fields(id),
    value           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, field_id)
);
CREATE INDEX IF NOT EXISTS idx_ecfv_tenant ON employee_custom_field_values (tenant_id);
CREATE INDEX IF NOT EXISTS idx_ecfv_employee ON employee_custom_field_values (employee_id);
