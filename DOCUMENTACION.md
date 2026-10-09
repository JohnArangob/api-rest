# 📚 DOCUMENTACIÓN de la API de Libros (desde cero)

> **Para quién es esto:** para vos, que estás arrancando con APIs y Go desde
> nivel 0. No hace falta que sepas nada previo. Cada concepto se explica la
> primera vez que aparece, sin jerga sin definir.

---

## Índice

1. [¿Qué es una API?](#1--qué-es-una-api)
2. [¿Qué es una API REST?](#2--qué-es-una-api-rest)
3. [¿Qué es HTTP, un verbo y un código de estado?](#3--http-verbos-y-códigos-de-estado)
4. [¿Qué es JSON?](#4--qué-es-json)
5. [Qué hace exactamente ESTA API](#5--qué-hace-exactamente-esta-api)
6. [Los archivos del proyecto y qué hace cada uno](#6--los-archivos-del-proyecto)
7. [La arquitectura por capas: ¿POR QUÉ está así?](#7--la-arquitectura-por-capas--por-qué)
8. [Recorrido completo de una petición](#8--recorrido-completo-de-una-petición)
9. [Cómo ejecutarla y probarla](#9--cómo-ejecutarla-y-probarla)
10. [Anatomía de cada endpoint](#10--anatomía-de-cada-endpoint)
11. [El bug del `null` en tu curl (lección completa)](#11--el-bug-del-null)
12. [Conceptos de Go que usa este proyecto](#12--conceptos-de-go-que-usa-este-proyecto)
13. [Glosario](#13--glosario)
14. [Cosas para mejorar (sin aplicar todavía)](#14--cosas-para-mejorar)

---

## 1. 🤔 ¿Qué es una API?

**API** = *Application Programming Interface* = **Interfaz de Programación
de Aplicaciones**.

Suena aterrador, pero la idea es simple:

> Una API es un **medio para que dos programas se hablen entre sí**,
> siguiendo reglas claras.

No es una web, no es una base de datos, no es un servidor. Es el **contrato**
que dice: *"si me pedís esto por acá, yo te devuelvo esto otro"*.

### Ejemplo de la vida real

Pensá en una **máquina de café**:

- Vos le **mandás** un pedido (apretar botón "espresso").
- La máquina te **devuelve** un café (o un error si no hay agua).

La máquina es la API. Vos sos el cliente. El café es la **respuesta**.

En programación:

```
TU PROGRAMA (cliente)  ──pide──▶  LA API  ──responde──▶  TU PROGRAMA
   ej: curl                           ej: esta API             ves el JSON
```

---

## 2. 🌐 ¿Qué es una API REST?

**REST** (*Representational State Transfer*) es un **estilo** de diseñar APIs.
No es un lenguaje ni una librería: es un conjunto de buenas prácticas.

Las reglas principales de REST:

| Regla | Significado | Ejemplo en esta API |
|---|---|---|
| Todo es un **recurso** | Cada cosa que manejás tiene nombre | `books` (los libros) |
| Cada recurso tiene una **URL** | Una dirección para identificarlo | `/books/1` |
| Usás **verbos HTTP** para actuar | GET, POST, PUT, DELETE | ver tabla abajo |
| Es **sin estado** | Cada petición se entiende sola; el servidor no recuerda "quién sos" entre una y otra | no hay sesión/login |
| Los datos van en **JSON** | El formato estándar de intercambio | `{"title": "..."}` |

**¿Por qué REST y no otra cosa?** Porque es simple, está soportado por todo
(navegadores, curl, Postman, móviles) y no requiere instalar nada especial.
Para una API de libros, es la opción natural.

---

## 3. 🚦 HTTP, verbos y códigos de estado

**HTTP** es el lenguaje que usan los clientes y los servidores para hablar.
Una petición HTTP tiene, como mínimo:

```
VERBO   RUTA            CUERPO (opcional)
PUT     /books/1        {"title":"...","author":"..."}
```

### Los verbos (métodos) que usa esta API

| Verbo | Significado | ¿Crea? | ¿Trae datos? | ¿Borra? |
|---|---|---|---|---|
| **GET** | Dame este recurso | no | **sí** | no |
| **POST** | Creá un recurso nuevo | **sí** | no | no |
| **PUT** | Actualizá este recurso | no | no | no |
| **DELETE** | Borrá este recurso | no | no | **sí** |

> Nota: en esta API los tres verbos que tocan `/books/{id}` (GET, PUT,
> DELETE) viven en el MISMO handler, y se separan con un `switch r.Method`.

### Los códigos de estado (los "códigos de salida")

Son números que el servidor manda junto a la respuesta para decir **cómo
salió la cosa**:

| Código | Nombre | Cuándo se usa en esta API |
|---|---|---|
| `200` | OK | GET exitoso (implícito: no se llama a `WriteHeader`) |
| `201` | Created | POST que creó un libro |
| `204` | No Content | DELETE exitoso (borraste, no hay nada que devolver) |
| `400` | Bad Request | JSON inválido o id de la URL que no es número |
| `404` | Not Found | No existe el libro con ese id |
| `405` | Method Not Allowed | Mandaste un método que esa ruta no soporta |
| `500` | Internal Server Error | Falló algo de mi lado (la base de datos, por ej.) |

---

## 4. 📦 ¿Qué es JSON?

**JSON** (*JavaScript Object Notation*) es el formato de texto con el que se
pasan datos entre un cliente y una API. Se ve así:

```json
{
  "id": 1,
  "title": "Las aventuras de Alexis y fido",
  "author": "Alex Jiminix"
}
```

Reglas básicas:

- Los **nombres de clave** van entre comillas dobles.
- Se usan `{ }` para objetos y `[ ]` para listas.
- Valores pueden ser texto (entre comillas), números, `true`/`false`, `null`.

### Relación con el código de Go

En `internal/model/book.go` tenés esto:

```go
type Libro struct {
    ID     int    `json:"id"`
    Titulo string `json:"title"`
    Autor  string `json:"author"`
}
```

Esos tags `json:"..."` son el **puente** entre el nombre en Go (`Titulo`) y
el nombre en JSON (`title`). Sin ellos, la API respondería `"Titulo"` y tu
curl con `"title"` no calzaría.

---

## 5. 📖 Qué hace exactamente ESTA API

Es una **API CRUD de libros**.

**CRUD** = **C**reate, **R**ead, **U**pdate, **D**elete (Crear, Leer,
Actualizar, Borrar). Es el patrón mínimo de cualquier API que guarda cosas.

Datos de la API:

- **Recurso:** libros (`books`)
- **Campos:** `id` (número, lo pone la BD), `title` (texto), `author` (texto)
- **Base de datos:** SQLite, guardada en el archivo `books.db` (un archivo
  solo, no hace falta instalar un servidor de BD)
- **Puerto:** `8080`

Acciones disponibles:

| Acción | Verbo | Ruta | Qué hace |
|---|---|---|---|
| Listar todos | GET | `/books` | devuelve todos los libros |
| Crear | POST | `/books` | crea un libro y devuelve su id |
| Ver uno | GET | `/books/{id}` | devuelve un libro |
| Actualizar | PUT | `/books/{id}` | cambia title y author |
| Borrar | DELETE | `/books/{id}` | lo elimina |

---

## 6. 🗂️ Los archivos del proyecto

```
api-rest/
├── go.mod                          ← qué es este módulo y qué dependencias tiene
├── go.sum                          ← "huella digital" de las dependencias (generado)
├── books.db                        ← la base de datos (archivo; se crea sola)
├── README.md
├── DOCUMENTACION.md                ← este archivo
├── main.go                         ← PUNTO DE ENTRADA: arma todo y enciende el servidor
└── internal/                       ← "internal" = código interno, que otros
    │                                   proyectos no pueden importar
    ├── model/book.go               ← QUÉ datos manejamos (el molde del Libro)
    ├── service/book_service.go     ← QUÉ reglas se aplican (lógica de negocio)
    ├── store/book_store.go         ← CÓMO se guardan (SQL / base de datos)
    └── transport/book_handler.go   ← CÓMO se recibe y responde (HTTP/JSON)
```

### Qué significa "capa" y por qué hay 4 carpetas

Una **capa** es un grupo de código con una única responsabilidad. Acá:

| Capa | Responsabilidad | NO le corresponde |
|---|---|---|
| `transport` | Recibir HTTP, responder HTTP | Escribir SQL, validar reglas |
| `service` | Aplicar reglas de negocio | Escribir HTTP, escribir SQL |
| `store` | Ejecutar SQL | Validar reglas, saber de HTTP |
| `model` | Definir la forma de los datos | Nada más (no tiene lógica) |

---

## 7. 🏗️ La arquitectura por capas: ¿POR QUÉ?

Esta es **la pregunta más importante** de todo el documento. Casi todo el
resto es detalle técnico; esto es diseño.

### La dependencia (quién llama a quieta)

```
        main.go (arma todo)
            │
            ▼
    ┌───────────────┐
    │  transport     │  ← habla HTTP y JSON
    │  (book_handler)│
    └───────┬───────┘
            │ llama a
            ▼
    ┌───────────────┐
    │  service       │  ← aplica reglas ("el título no puede estar vacío")
    │  (book_service)│
    └───────┬───────┘
            │ llama a
            ▼
    ┌───────────────┐
    │  store         │  ← ejecuta SQL
    │  (book_store)  │
    └───────┬───────┘
            │
            ▼
       books.db (SQLite)
```

Las flechas **solo van para abajo**. Ni `store` conoce a `transport`, ni
`service` sabe que existe HTTP.

### ¿Y si NO hago esto? (el "monolito")

Verías todo junto en `main.go`:

```go
// EJEMPLO de lo que NO hay que hacer (no está en tu proyecto)
if r.Method == "PUT" {
    if body == "" { ... }                       // regla de negocio
    db.Exec("UPDATE books SET ...")             // SQL
    w.Write(json)                               // HTTP
}
```

Funciona al principio, pero:

- **No podés probar** las reglas sin levantar un servidor HTTP.
- **No podés cambiar** la base de datos sin reescribir todo.
- **No podés reusar** las mismas reglas si mañana agregás una app móvil.
- Cualquier cambio toca el mismo archivo gigante → miedo a romper todo.

### ¿Y si lo separo? (lo que hiciste vos)

- Cambiás SQLite por PostgreSQL → solo tocás `store`.
- Agregás otra regla (ej: "el autor no puede estar vacío") → solo tocás `service`.
- Cambiás de HTTP a gRPC → solo tocás `transport`.
- Tus reglas de negocio quedan **reutilizables y testeables**.

> **En una frase:** la arquitectura por capas existe para que un cambio en
> un lado **no rompa** el otro.

### La "inyección de dependencias" (lo del PASO 3 en `main.go`)

```go
bookStore := store.New(db)                  // la capa de datos recibe la BD
bookService := service.New(bookStore)       // la de reglas recibe la de datos
bookHandler := transport.New(bookService)   // la de HTTP recibe la de reglas
```

Cada capa recibe sus dependencias **por parámetro**, desde afuera. ¿Para qué?

- Hoy `service` recibe un `store` que usa SQLite.
- Mañana podés pasarle un store que use PostgreSQL (o uno falso para tests)
  y el código de `service` **no cambia en absoluto**.

Es como enchufar un lavarropas en la pared: el lavarropas no decide dónde
toma el agua, vos le pasás la manguera.

### ¿Por qué `model` no depende de nadie?

Porque `model` solo dice *qué campos tiene un libro*. No sabe que existe SQL,
ni HTTP, ni reglas. Eso permite que las **tres** capas lo usen sin ciclos
(A depende de B y B de A, que en programación es un problema clásico).

### ¿Por qué `internal/`?

En Go, la carpeta `internal/` marca código **privado del módulo**: si alguien
importa tu proyecto, **no podrá** usar nada de adentro de `internal/`. Es una
forma de decir "esto es detalle interno, la cara pública es la API HTTP".

---

## 8. 🚶 Recorrido completo de una petición

Vamos a seguir tu propio curl, paso por paso, viendo en qué archivo se
detiene cada cosa.

```bash
curl -X PUT -H "Content-Type: application/json" \
     -d '{"title":"Las aventuras de Alexis y fido","author":"Alex Jiminix"}' \
     http://localhost:8080/books/1
```

Traducido a humano:

- `-X PUT` → "voy a usar el método PUT".
- `-H "Content-Type: application/json"` → "el cuerpo que mando es JSON".
- `-d '{...}'` → "este es el cuerpo (body) de la petición".
- `http://localhost:8080/books/1` → "al servidor en mi máquina, puerto 8080,
  recurso `/books/1`".

### Los 6 pasos

```
PASO 1 ─ main.go:43
  http.HandleFunc("/books/", bookHandler.HandleBookByID)
  El repartidor de rutas (mux) ve que la URL empieza con "/books/"
  y le pasa la petición a HandleBookByID.

PASO 2 ─ internal/transport/book_handler.go : HandleBookByID
  2a. Saca el id de la URL:  strings.TrimPrefix("/books/1", "/books/") → "1"
  2b. Lo convierte a número: strconv.Atoi("1") → 1
  2c. Ve que r.Method es PUT → entra al case http.MethodPut
  2d. Lee el body JSON y lo volca en un struct model.Libro
      {"title":"...", "author":"..."} → Libro{Titulo: "...", Autor: "..."}

PASO 3 ─ internal/service/book_service.go : UpdateAlLibro
  Aplica la REGLA DE NEGOCIO:
      if libro.Titulo == "" → error "Necesitamos el titulo"
  Como el título no está vacío, sigue.

PASO 4 ─ internal/store/book_store.go : Update
  Ejecuta el SQL:
      UPDATE books SET title = ?, author = ? WHERE id = 1
  Y devuelve el puntero al libro con los datos nuevos.

PASO 5 ─ de vuelta en el handler
  json.NewEncoder(w).Encode(updated)  →  convierte el struct en JSON
  Se escribe en la respuesta, con Content-Type: application/json.

PASO 6 ─ curl
  Devuelve a tu consola:
  {"id":1,"title":"Las aventuras de Alexis y fido","author":"Alex Jiminix"}
```

**Clave:** el dato sube `transport → service → store` y la respuesta vuelve
`store → service → transport`. Cada capa solo habla con la de al lado.

---

## 9. ▶️ Cómo ejecutarla y probarla

### 1) Levantar el servidor

```bash
go run .
```

Si todo está bien, en la consola ves:

```
Servidor ejecutandose en http:localhots:8080
API Endpoints:
...
```

Y el programa **queda esperando** (no termina). Para cortarlo: `Ctrl + C`.

> Dato: si `books.db` no existe, se crea sola al arrancar, con la tabla
> `books` adentro.

### 2) Probar los 5 endpoints

Abrí **otra terminal** (la del servidor sigue ocupada) y ejecutá:

```bash
# CREAR un libro (respuesta 201 + el libro con id)
curl -X POST -H "Content-Type: application/json" \
     -d '{"title":"El principito","author":"Saint-Exupery"}' \
     http://localhost:8080/books

# LISTAR todos (respuesta 200 + array de libros)
curl http://localhost:8080/books

# VER uno solo
curl http://localhost:8080/books/1

# ACTUALIZAR
curl -X PUT -H "Content-Type: application/json" \
     -d '{"title":"Titulo nuevo","author":"Autor nuevo"}' \
     http://localhost:8080/books/1

# BORRAR (respuesta 204, sin cuerpo)
curl -i -X DELETE http://localhost:8080/books/1
```

**Tip:** agregá `-i` a cualquier curl para ver los headers y el código de
estado (200, 201, 404...). Sin `-i` solo ves el cuerpo, y no sabés qué
código devolvió.

```bash
curl -i http://localhost:8080/books
# verás algo como:
# HTTP/1.1 200 OK
# Content-Type: application/json
# ...
```

### 3) Probar casos de error (¡esto es lo que más enseña!)

```bash
# JSON roto → debería dar 400
curl -X PUT -d 'esto no es json' http://localhost:8080/books/1

# Título vacío → la regla del service lo rechaza → 500
curl -X PUT -H "Content-Type: application/json" \
     -d '{"title":"","author":"X"}' http://localhost:8080/books/1

# id que no es número → 400
curl http://localhost:8080/books/hola

# id inexistente → 404
curl http://localhost:8080/books/9999

# Método que no existe en esa ruta → 405
curl -X PATCH http://localhost:8080/books
```

---

## 10. 🧩 Anatomía de cada endpoint

### `GET /books` → listar todos

- **Archivo:** `transport/book_handler.go`, `HandleBooks`, `case http.MethodGet`
- **Service:** `ObtenTodosLosLIbros`
- **Store:** `GetAll` → `SELECT id, title, author FROM books`
- **Código:** 200 (implícito)
- **Respuesta:** array de libros, o `null` si la tabla está vacía
  (Go codifica un slice vacío/`nil` como `null`, por eso a veces ves `null` en
  el GET cuando todavía no creaste nada).

### `POST /books` → crear

- **Archivo:** `HandleBooks`, `case http.MethodPost`
- **Service:** `CrearLibro` (hoy no valida nada)
- **Store:** `Create` → `INSERT INTO books (title, author) VALUES (?, ?)`
  y luego `LastInsertId()` para saber el id nuevo
- **Código:** 201 Created
- **Respuesta:** el libro con su `id` recién asignado

### `GET /books/{id}` → ver uno

- **Archivo:** `HandleBookByID`, `case http.MethodGet`
- **Store:** `GetByID` → `SELECT ... WHERE id = ?` con `QueryRow`
- **Códigos:** 200 si existe, 404 si no

### `PUT /books/{id}` → actualizar

- **Archivo:** `HandleBookByID`, `case http.MethodPut`
- **Service:** `UpdateAlLibro` ← **acá vive la única regla de negocio**
  del proyecto: el título no puede estar vacío
- **Store:** `Update` → `UPDATE books SET title = ?, author = ? WHERE id = ?`
- **Códigos:** 200, 400 (JSON malo), 500 (título vacío u error de BD)

### `DELETE /books/{id}` → borrar

- **Archivo:** `HandleBookByID`, `case http.MethodDelete`
- **Store:** `Delete` → `DELETE from books WHERE id = ?`
- **Código:** 204 No Content, sin cuerpo

---

## 11. 🐛 El bug del `null`

### Qué pasaba

```bash
curl -X PUT ... http://localhost:8080/books/1
null        ← en vez de el libro actualizado
```

### Por qué salía `null`

La cadena era:

1. El store terminaba con `return nil, nil` → es decir: **"no hay error, pero
   tampoco te devuelvo el libro"**.
2. El handler recibía `updated = nil` y hacía
   `json.NewEncoder(w).Encode(updated)`.
3. Go, al serializar un puntero `nil`, escribe literalmente la palabra
   **`null`** en JSON (que es el "no hay nada" de JSON).

Es decir: **el `null` no venía de tu curl, ni de tu JSON, ni de la URL**.
Venía de la función `Update` que no devolvía el puntero.

### Regla general para recordar

> Cuando una API te devuelve `null`, la primera sospecha debe ser:
> **¿qué estoy devolviendo?** casi siempre es un puntero/slice sin inicializar
> que llega al `Encode`.

### El segundo detalle (el que hace que no se guarde)

En `Update`, el `WHERE` usa el parámetro `id` (el de la URL), y no `libro.ID`
(el que mandó el cliente en el body). Esto es **correcto**:

- La URL manda sobre **qué fila** modificar → `/books/1` actualiza el id 1.
- El body manda sobre **qué campos** cambian → title y author.

Si se usara `libro.ID`, dependerías de que el cliente mande también el `id`
en el JSON (y no lo manda, o manda otro), con lo que el `WHERE` no matchearía
ninguna fila: la API respondería 200 "todo bien" pero **no se habría
cambiado nada en la base**. Es un bug clásico y muy difícil de ver, porque
**no da error**.

---

## 12. 🧱 Conceptos de Go que usa este proyecto

### Paquetes (`package`)

Un paquete es una carpeta con funciones relacionadas. `main` es el paquete
especial que Go necesita para generar el ejecutable. Cada carpeta de
`internal/` es su propio paquete (`model`, `service`, `store`, `transport`).

Para usar algo de otro paquete, se importa:

```go
import "api-rest/internal/model"
```

El nombre del import (`api-rest/...`) sale de `go.mod`, que dice
`module api-rest`.

### El puntero (`*` y `&`)

```go
*model.Libro   // "puntero a un Libro": una flecha que señala a un Libro
&libro         // "dirección de libro": dónde está ese Libro en memoria
```

¿Para qué? Para **no copiar** el struct entero y para poder **modificarlo
desde otra función**. Si le pasás `libro` (sin `&`), la otra función recibe
una copia y lo que cambie ahí no le llega al original.

### Los errores son valores, no excepciones

En Go **no hay** `try/catch`. Se hace esto:

```go
resultado, err := algoQuePuedeFallar()
if err != nil {
    return nil, err   // corto y paso el error para arriba
}
```

Cada capa decide qué hacer con el error: `store` lo devuelve tal cual,
`service` lo propaga, y `transport` lo traduce en un código HTTP
(404, 500...).

### Interfaces

```go
type Store interface {
    GetAll() ([]*model.Libro, error)
    ...
}
```

Es una **lista de promesas**: "cualquier cosa que tenga estos métodos, para
mí es un `Store`". Le da flexibilidad: el `service` no sabe si por dentro
hay SQLite o una lista en memoria, solo sabe que puede pedirle `GetAll()`.

### `defer`

```go
defer db.Close()
```

"Ejecutá esto cuando la función termine, pase lo que pase". Se usa para
tareas de limpieza (cerrar conexiones, archivos, etc.).

### `nil`

El "nulo" de Go: indica que un puntero/slice/interface no apunta a nada.
`nil` sin error = "salió bien pero no tengo nada que darte" → y eso en JSON
se imprime como `null`.

---

## 13. 📖 Glosario

| Término | Significado corto |
|---|---|
| **API** | Contrato para que dos programas se hablen |
| **REST** | Estilo de diseño de APIs con reglas fijas |
| **HTTP** | Protocolo de comunicación cliente–servidor |
| **Verbo / método** | GET, POST, PUT, DELETE: la acción a realizar |
| **Código de estado** | 200, 404, 500... resultado de la petición |
| **Endpoint** | Una ruta + un método (ej: `PUT /books/1`) |
| **JSON** | Formato de texto para pasar datos |
| **CRUD** | Create, Read, Update, Delete |
| **Capa** | Grupo de código con una única responsabilidad |
| **Dependencia** | Pieza que otra necesita para funcionar |
| **Inyección de dependencias** | Pasarle a cada pieza lo que necesita desde afuera |
| **Modelo** | Struct que define la forma de los datos |
| **Handler** | Función que atiende una petición HTTP |
| **Store / repositorio** | Capa que habla con la base de datos |
| **Service** | Capa con la lógica/reglas de tu app |
| **Puntero** | Dirección en memoria de un valor |
| **Interface** | Lista de promesas (métodos) que un tipo cumple |
| **Placeholder (`?`)** | Marcador en SQL para pasar valores seguros |
| **Puerto** | "Ventana" del servidor donde escucha (8080) |
| **Mux** | Repartidor de rutas: URL → handler |

---

## 14. 🔧 Cosas para mejorar (sin aplicar todavía)

Estas son observaciones de diseño/código. **No se modificó nada**: las dejo
acá como guía para cuando quieras seguir aprendiendo.

1. **`POST` setea el header después de `WriteHeader`**
   En `HandleBooks`, el `case http.MethodPost` llama
   `w.WriteHeader(http.StatusCreated)` **antes** de `w.Header().Set(...)`.
   En Go, una vez enviado el estado, los headers posteriores **no se envían**.
   El orden correcto es: primero `Header().Set`, después `WriteHeader`.

2. **Falta un `return` tras el error de parseo de id**
   En `HandleBookByID`, si `strconv.Atoi` falla se responde 400 pero la
   función sigue ejecutándose con `id = 0`. Sumar el `return` evita el
   doble trabajo.

3. **`GET /books` con tabla vacía devuelve `null`**
   Un slice sin elementos se serializa como `null`. Lo ideal sería devolver
   `[]` (array vacío), que es más claro para quien consume la API.

4. **`Delete` no verifica si existía el id**
   Borrar un id inexistente también responde 204. Se podría usar
   `RowsAffected()` para distinguir "borrado" de "no existía" (404).

5. **Validaciones en `CrearLibro`**
   Hoy el service solo valida el título en el PUT. El POST crea libros con
   título o autor vacíos sin quejarse.

6. **Mensajes de error en inglés y español mezclados**
   Conviene unificar el idioma de las respuestas.

7. **El puerto y la ruta de la BD están "hardcodeados"**
   En producción se usan variables de entorno (`os.Getenv`) para poder
   cambiarlos sin recompilar.

8. **No hay tests**
   Justo la ventaja de las capas es que `service` se puede testear sin
   servidor HTTP ni base de datos real. Es el siguiente gran paso de
   aprendizaje (`go test ./...`).

---

## ✅ Resumen para tenerlo de memoria

```
main.go         → arma y enciende todo
transport       → HTTP  (entrar/salir)   → traduce a llamadas y a JSON
service         → reglas (qué está bien) → valida antes de guardar
store           → SQL    (guardar)       → solo habla con la BD
model           → forma   (qué campos)   → no depende de nada

Una petición baja  transport → service → store
La respuesta sube  store → service → transport

Los verbos:  GET traer · POST crear · PUT actualizar · DELETE borrar
Los códigos: 200 ok · 201 creado · 204 sin contenido · 400 malo ·
             404 no existe · 405 método mal · 500 error mío
El null:     siempre sale de un puntero nil llegando a json.Encode
```
