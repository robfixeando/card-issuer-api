# Card Issuer Operations API

Micro-API backend construida en **Go**, que implementa un sistema de operaciones de emisión de tarjetas (Challenge técnico para NovoPayment) con aislamiento estricto de multi-tenancy y cumplimiento de normativas de seguridad (PCI-DSS).

## 🚀 Stack Tecnológico

* **Lenguaje:** Go (Golang)
* **Framework Web:** Gin
* **ORM:** GORM
* **Base de datos:** SQLite (utilizando el driver puro en Go `github.com/glebarez/sqlite` para evitar dependencias de CGO/gcc y fricciones de compilación)
* **Pruebas:** Testing nativo de Go y `net/http/httptest`

## 🔒 Decisiones de Arquitectura y Seguridad

1. **Multi-Tenancy Estricto:**
   * Garantizado mediante un middleware HTTP que valida el header `X-Tenant-ID`.
   * Todas las consultas a la base de datos (`Create`, `Find`, `First`, `Save`) filtran automáticamente por `tenant_id`, asegurando que ningún banco o tenant pueda acceder a los datos de otro.

2. **Cumplimiento PCI-DSS (Minimización de Datos):**
   * **Cero almacenamiento de PAN en texto plano:** Los números de tarjeta (PAN) nunca se guardan completos.
   * **Hash SHA-256:** Se almacena un hash criptográfico (`pan_hash`) para validaciones y auditoría.
   * **Últimos 4 dígitos:** Solo se guardan los últimos 4 dígitos (`last4`) para identificación y soporte al tarjetahabiente.

3. **Procesamiento por Lotes Atómico:**
   * La emisión masiva de tarjetas utiliza transacciones de base de datos de GORM (`db.Transaction`) para garantizar atomicidad: si falla una sola tarjeta del lote, se revierte toda la operación.

## 🛠️ Instalación y Configuración

### 1. Prerrequisitos

Asegúrate de tener **Go** instalado en tu máquina.

### 2. Clonar y Configurar

```bash
# Navegar al directorio del proyecto
cd ~/proyectos/card-issuer-api

# Inicializar y descargar dependencias
go mod tidy
```

### 3. Ejecutar el Servidor

```bash
go run main.go
```

El servidor se iniciará localmente en el puerto **`http://localhost:8085`**.

## 🧪 Ejecución de Pruebas

Ejecuta la suite completa de pruebas unitarias e integración (utilizando una base de datos SQLite en memoria):

```bash
go test -v ./...
```

## 📡 Endpoints de la API y Ejemplos de Uso

Todas las peticiones requieren obligatoriamente el header `X-Tenant-ID` para identificar a la institución emisora.

### 1. Emitir una Tarjeta (Individual)

* **Endpoint:** `POST /api/v1/cards`
* **Header:** `X-Tenant-ID: banco-galicia`
* **Comando cURL:**

```bash
curl -X POST http://localhost:8085/api/v1/cards \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: banco-galicia" \
  -d '{
    "cardholder_name": "Juan Perez",
    "pan": "4532111122223333"
  }'
```

### 2. Emitir Tarjetas por Lotes (Atómico)

* **Endpoint:** `POST /api/v1/cards/batch`
* **Header:** `X-Tenant-ID: banco-galicia`
* **Comando cURL:**

```bash
curl -X POST http://localhost:8085/api/v1/cards/batch \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: banco-galicia" \
  -d '{
    "cards": [
      {"cardholder_name": "Ana Gomez", "pan": "4532555566667777"},
      {"cardholder_name": "Carlos Ruiz", "pan": "4532888899990000"}
    ]
  }'
```

### 3. Listar Tarjetas (con Paginación y Filtro por Estado)

* **Endpoint:** `GET /api/v1/cards?page=1&limit=10&status=ISSUED`
* **Header:** `X-Tenant-ID: banco-galicia`
* **Comando cURL:**

```bash
curl -X GET "http://localhost:8085/api/v1/cards?page=1&limit=10&status=ISSUED" \
  -H "X-Tenant-ID: banco-galicia"
```

### 4. Consultar Detalle de una Tarjeta por ID

* **Endpoint:** `GET /api/v1/cards/{id}`
* **Header:** `X-Tenant-ID: banco-galicia`
* **Comando cURL:**

```bash
curl -X GET http://localhost:8085/api/v1/cards/1 \
  -H "X-Tenant-ID: banco-galicia"
```

### 5. Actualizar Estado de una Tarjeta (Bloquear / Revocar)

* **Endpoint:** `PATCH /api/v1/cards/{id}/status`
* **Header:** `X-Tenant-ID: banco-galicia`
* **Estados válidos:** `ISSUED`, `BLOCKED`, `REVOKED`
* **Comando cURL:**

```bash
curl -X PATCH http://localhost:8085/api/v1/cards/1/status \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: banco-galicia" \
  -d '{"status": "BLOCKED"}'