# Decision Log & Tradeoff Write-up - Card Issuer API

## 1. Decisiones de Arquitectura
- **Estructura directa en Handlers:** Decidí no armar capas de servicios o repositorios para mantener el código simple y fácil de leer.
- **SQLite sin CGO:** Usé el driver puro en Go.

## 2. Estrategia para Procesamiento por Lotes (Batch Processing)
- **Transacciones atómicas:** Todo el lote corre dentro de una transacción de GORM. Si una tarjeta del lote viene mal o falla, se hace `rollback` automático de todo el bloque para no dejar la base de datos a medias.

## 3. Seguridad y Cumplimiento (PCI-DSS)
- **Manejo del PAN:** El número de tarjeta nunca se guarda en texto plano. Guardamos un hash SHA-256 (`pan_hash`) para identificarlo y solo los últimos 4 dígitos (`last4`) para mostrar en pantalla.
- **Gaps y pendientes para producción:**
  - Si en un escenario real se necesitara recuperar el número completo, faltaría implementar un cifrado simétrico (como AES-256) con llaves rotativas en lugar de un hash unidireccional.
  - Habría que sumar un rate limiter por tenant y logs de auditoría más estrictos para los cambios de estado críticos.