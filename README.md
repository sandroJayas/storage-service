# Storage Service



## 🚀 Getting Started

### Run locally

```
docker compose up
task run
```

### Run integration test

```
docker compose up
task run
task test
```

### Run migrations
Migration files live in /migrations/sql

To apply them and regenerate the schema use:
```
task regenerate-schema
```

### Run swagger

when the controller is changed with additional annotations, to update swagger docs run

```
task swagger
```
to see the docs, run:
```
docker compose up
task run
```
Then go to http://localhost:8080/swagger/index.html