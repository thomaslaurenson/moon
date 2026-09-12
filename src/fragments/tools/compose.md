# Docker Compose conventions

Conventions for `docker-compose.yml`. Assumes the Docker fragment for the Dockerfile each service builds from.

## Structure

Each service lives in its own directory containing a `Dockerfile` and, when needed, a `.dockerignore`. The location of service directories depends on the project type:

**Standalone docker or infrastructure project**: service directories at the project root:

```text
api/
  Dockerfile
postgres/
  Dockerfile
docker-compose.yml
```

**Monorepo**: service directories under a `docker/` folder:

```text
docker/
  api/
    Dockerfile
  postgres/
    Dockerfile
src/
docker-compose.yml
```

`docker-compose.yml` always lives at the project root.

## Build context

Every service must use a Dockerfile with an explicit build context. Never use the `image:` key directly; even for unmodified third-party images. This is a hard rule:

```yaml
# Good - always use a Dockerfile
services:
  postgres:
    build:
      context: ./postgres
      dockerfile: Dockerfile

# Bad - never reference an image directly
services:
  postgres:
    image: postgres:16.2-alpine
```

A Dockerfile for an unmodified third-party image contains only the `FROM` line until customisation is needed:

```dockerfile
FROM postgres:16.2-alpine
```

## Version field

Never include the `version:` field. It is deprecated in Docker Compose V2 and must not be added:

```yaml
# Good
services:
  api:
    build:
      context: ./api
      dockerfile: Dockerfile

# Bad - version field is deprecated
version: "3.8"
services:
  api:
    ...
```

## Volumes

Use named volumes for all persistent data. Never use anonymous volumes; they are untrackable and difficult to manage. Always declare named volumes explicitly at the bottom of `docker-compose.yml`:

```yaml
services:
  postgres:
    build:
      context: ./postgres
      dockerfile: Dockerfile
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Networking

Single-service projects do not need explicit network configuration; Docker Compose provides a default network automatically.

Multi-service projects must define a named `backend` network for private inter-service communication. Never expose internal services directly to the host network. Always declare networks explicitly at the bottom of `docker-compose.yml` alongside volumes:

```yaml
services:
  api:
    build:
      context: ./api
      dockerfile: Dockerfile
    networks:
      - backend

  postgres:
    build:
      context: ./postgres
      dockerfile: Dockerfile
    networks:
      - backend

networks:
  backend:
    driver: bridge

volumes:
  postgres_data:
```
