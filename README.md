# TaskFlow API (Go)

![CI](https://github.com/mbeltran93/taskflow-go/actions/workflows/ci.yml/badge.svg)

API de gestion de tareas/proyectos (version reducida de Trello/Jira), escrita en
Go puro + PostgreSQL. Es el miembro Go de un portafolio de backends que
implementan el mismo dominio en distintos stacks; este repo existe para
mostrar Go como complemento del stack principal (Java 21 / Spring Boot / GCP),
porque Go es muy comun en el ecosistema cloud-native de GCP (Cloud Run, gRPC,
microservicios livianos, herramientas del propio Kubernetes).

## Por que Go, en un perfil de GCP/Cloud Run

- **Imagenes minusculas**: el binario final es estatico (`CGO_ENABLED=0`) y la
  imagen Docker se construye sobre `scratch` - sin JVM, sin libc, sin shell.
  La imagen de este proyecto pesa **~12 MB** (ver mas abajo), contra cientos
  de MB tipicas de una imagen basada en JRE. En Cloud Run eso es arranque en
  frio mas rapido y menos superficie de ataque.
- **Concurrencia liviana**: goroutines y el `net/http` estandar alcanzan para
  un servicio de este tamano sin un framework pesado encima.
- **Sin runtime que mantener**: el binario compilado es el deploy entero;
  no hay que elegir heap size, GC, ni imagen base con una version de JVM.
- Es el lenguaje en el que esta escrito buena parte del ecosistema que un
  perfil GCP/K8s ya opera (`kubectl`, `terraform`, `docker`, el propio
  `Cloud Run`), asi que leer y escribir Go facilita trabajar con ese
  ecosistema aunque el dia a dia sea Java.

## Stack

| Capa | Eleccion | Por que |
|---|---|---|
| Router HTTP | [`chi`](https://github.com/go-chi/chi) | Minimalista e idiomatico: es `net/http` con un router encima, no un framework que impone su propia forma de escribir handlers. Para un portafolio que quiere mostrar Go "puro", pesa mas eso que las comodidades de gin. |
| Acceso a datos | [`sqlx`](https://github.com/jmoiron/sqlx) sobre `database/sql` + [`pgx`](https://github.com/jackc/pgx) | Se evaluo `sqlc` (SQL type-safe generado, similar en espiritu a jOOQ) pero requiere una toolchain de codegen adicional; para el alcance de este proyecto, `sqlx` da binding a structs sin ORM y sin ese paso de build extra, manteniendo el SQL explicito a mano. `pgx` es el driver de PostgreSQL mas activo y rapido para Go. |
| Migraciones | [`golang-migrate`](https://github.com/golang-migrate/migrate) | Los `.sql` versionados viven embebidos en el binario (`go:embed`) y se aplican solos al arrancar - `docker compose up` deja la base lista sin un paso manual. |
| Auth | [`golang-jwt/jwt`](https://github.com/golang-jwt/jwt) + `bcrypt` (`golang.org/x/crypto`) | JWT HS256 con expiracion; passwords nunca se guardan en texto plano. |
| Docs | OpenAPI 3 a mano (`openapi.yaml`) + Swagger UI servido por la propia app | Sin generar nada: la app embebe el YAML (`go:embed`) y lo sirve en `/openapi.yaml`, con una pagina `/swagger` que carga `swagger-ui-dist` desde CDN. |
| Tests | `testing` estandar, tablas de casos con `t.Run` + [Testcontainers-go](https://golang.testcontainers.org/) | Unitarios contra fakes en memoria (sin red, sin Docker); integracion contra un Postgres real y efimero. |

## Arquitectura

Capas separadas por responsabilidad, sin framework de inyeccion de
dependencias - todo se conecta a mano con funciones constructoras desde
`cmd/api/main.go`:

```
cmd/api/main.go           - arranque: carga config, conecta DB, corre migraciones, wiring, servidor HTTP
internal/model/           - entidades de dominio (User, Project, Task, Status, Date)
internal/apperror/        - errores sentinela compartidos (ErrNotFound, ErrConflict, ErrInvalidInput, ErrUnauthorized)
internal/repository/      - interfaces de persistencia (lo que el service layer conoce)
internal/repository/postgres/ - implementacion real contra PostgreSQL (sqlx + pgx)
internal/service/         - logica de negocio; depende solo de las interfaces de repository (testeable con fakes)
internal/handler/         - capa HTTP: decodifica requests, llama al service, mapea errores a status codes
internal/middleware/      - JWT (RequireAuth) y logging estructurado con requestId
internal/auth/            - JWT y bcrypt
internal/config/          - lectura de variables de entorno
internal/db/              - conexion y migraciones (migraciones embebidas con go:embed)
internal/openapispec/     - openapi.yaml embebido en el binario
migrations/               - copia de las migraciones para usar golang-migrate CLI manualmente si hace falta
test/integration/         - tests end-to-end contra Postgres real (Testcontainers-go)
```

Go no tiene excepciones: los errores son valores. Los `service` devuelven
errores envueltos con `fmt.Errorf("...: %w", apperror.ErrNotFound)` y los
`handler` los traducen a status HTTP con `errors.Is` - el equivalente Go de
un `@ExceptionHandler` centralizado, sin runtime de excepciones.

La inyeccion de dependencias es, literalmente, pasar interfaces a funciones
constructoras (`NewUserService(userRepo)`). Los tests unitarios del paquete
`service` usan fakes en memoria que implementan esas mismas interfaces
(`internal/service/fakes_test.go`) en vez de un framework de mocking.

## Contrato de API

| Metodo | Ruta | Auth | Descripcion |
|---|---|---|---|
| POST | `/api/auth/login` | - | `{email, password}` -> `{token}` |
| GET | `/api/users` | - | Listar usuarios |
| POST | `/api/users` | - | Registrar usuario (no forma parte del contrato original del portafolio, pero sin registro no habria forma de crear cuentas; es publico a proposito, igual que en los repos hermanos) |
| GET | `/api/users/{id}` | - | Obtener usuario |
| GET | `/api/projects` | - | Listar proyectos |
| POST | `/api/projects` | JWT | Crear proyecto |
| GET | `/api/projects/{id}` | - | Obtener proyecto |
| PUT | `/api/projects/{id}` | JWT | Actualizar proyecto |
| DELETE | `/api/projects/{id}` | JWT | Borrar proyecto |
| GET | `/api/tasks?projectId=&status=` | - | Listar tareas, filtros opcionales |
| POST | `/api/tasks` | JWT | Crear tarea (status inicial `TODO`) |
| GET | `/api/tasks/{id}` | - | Obtener tarea |
| PUT | `/api/tasks/{id}` | JWT | Actualizar tarea |
| PATCH | `/api/tasks/{id}/status` | JWT | Cambiar status (`TODO`\|`IN_PROGRESS`\|`DONE`) |
| DELETE | `/api/tasks/{id}` | JWT | Borrar tarea |

Las rutas de escritura requieren `Authorization: Bearer <token>`. El
`openapi.yaml` completo (servido en `/openapi.yaml`, con UI en `/swagger`)
documenta cada request/response.

## Como correrlo

```bash
docker compose up --build
```

Eso levanta Postgres, corre las migraciones automaticamente al arrancar la
app, y publica la API en `http://localhost:8080`. Un flujo de punta a punta:

```bash
# registrar un usuario (necesario para poder loguearse despues)
curl -X POST localhost:8080/api/users -H "Content-Type: application/json" \
  -d '{"name":"Owner","email":"owner@example.com","password":"supersecret"}'

# login
curl -X POST localhost:8080/api/auth/login -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com","password":"supersecret"}'

# crear un proyecto con el token
curl -X POST localhost:8080/api/projects -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"Launch","description":"Q4","ownerId":1}'
```

Swagger UI: `http://localhost:8080/swagger`.

### Desarrollo local sin Docker

```bash
export DATABASE_URL=postgres://taskflow:taskflow@localhost:5432/taskflow?sslmode=disable
export JWT_SECRET=dev-secret
go run ./cmd/api
```

### Tests

```bash
go test ./internal/...     # unitarios, con fakes en memoria - no necesitan Docker
go test ./test/...         # integracion - levantan un Postgres real con Testcontainers-go
```

### Coleccion de Postman

`postman_collection.json` cubre el flujo completo (registro, login, crear
proyecto/tarea, cambiar status, filtrar, borrar) y los casos de error (401
sin token, 404 inexistente, 409 email duplicado). Validada con Newman:

```bash
docker run --rm --network host -v "$(pwd):/etc/newman" postman/newman:latest \
  run postman_collection.json --env-var baseUrl=http://localhost:8080
```

## Imagen Docker

Build multi-stage: `golang:1.26-alpine` para compilar un binario estatico
(`CGO_ENABLED=0`, `-ldflags="-s -w"`), copiado a una imagen `scratch` final
que solo contiene el binario y los certificados TLS.

```
taskflow-go:latest   12.3MB
```

## Limitaciones conocidas

- **Testcontainers dentro de un contenedor en Windows**: correr `go test
  ./test/...` directamente en Windows requiere tener Go instalado en el host
  (no fue el caso en esta maquina). Se verifico corriendo el test dentro de
  un contenedor `golang` con el socket de Docker Desktop montado y
  `--network host`; en Linux (incluido el runner de GitHub Actions en el CI
  de este repo) esto no hace falta, el proceso de test corre directo sobre
  el `docker.sock` del host sin la capa adicional de red.
- `ProjectRequest`/`TaskRequest` reciben `ownerId`/`assigneeId` del cliente en
  vez de inferirlos del JWT - es la misma convencion usada en los repos
  hermanos del portafolio para mantener el contrato idéntico entre stacks;
  en un sistema real el owner se derivaria del token, no del body.
- No hay paginacion en los listados (`GET /api/users`, `/api/projects`,
  `/api/tasks`); para el volumen de datos de un portafolio no hizo falta.
- El JWT no soporta refresh tokens ni revocacion; expira a los 60 minutos
  (configurable con `JWT_TTL_MINUTES`) y listo.
