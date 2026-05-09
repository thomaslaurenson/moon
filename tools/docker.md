# Docker Conventions

Conventions for Dockerfile and Docker Compose files across all projects.

## Design Principles

- Every service is built from a Dockerfile — never reference images directly in Compose
- Images are always pinned to a specific version — never use `latest` or untagged images
- Keep images as close to the official base as possible — avoid unnecessary packages
- Reproducibility over convenience — every build must produce the same result

---

## Images

### Pinning

Always pin images to the minor version. Never use `latest`, a major-only tag,
or an untagged reference:

```dockerfile
# Good
FROM alpine:3.20
FROM python:3.12-alpine
FROM postgres:16.2-alpine

# Bad
FROM alpine:latest
FROM alpine:3
FROM alpine
```

### Image preference

Select the base image in this order:

1. **Alpine** — default for all services. Minimal, small, and widely supported.
2. **Debian slim** — when Alpine's musl libc causes compatibility issues with
   C extensions or native libraries. Always add a comment explaining why Alpine
   was not used.
3. **`scratch`** — for precompiled static binaries. Zero OS overhead.
4. **Language-specific Alpine variants** — e.g. `python:3.12-alpine`,
   `golang:1.22-alpine` where the official image provides an Alpine base.

```dockerfile
# Deviating from Alpine — numpy requires glibc which is incompatible with musl
FROM python:3.12-slim
```

### Official images

Always prefer Docker Official Images (no namespace prefix). Use Vendor
Verified Publisher images only when no official alternative exists:

```
# Docker Official — always preferred
alpine, python, postgres, nginx, redis

# Vendor Verified Publisher — only when no official alternative exists
grafana/grafana, bitnami/postgresql
```

Never use unverified community images.

---

## Dockerfile

### Non-root user

Every Dockerfile must run the application as a non-root user. Check the
official image documentation first — many official images already provide
a non-root user (e.g. `postgres` provides the `postgres` user, `nginx`
provides `nginx`).

When no suitable user exists, create one named after the project or a
simplified version of it:

```dockerfile
RUN adduser -D -u 1001 myapp
USER myapp
```

For Debian-based images:

```dockerfile
RUN useradd -m -u 1001 myapp
USER myapp
```

The `USER` instruction must appear before `CMD` or `ENTRYPOINT`. This rule
does not apply to `scratch` images — there is no user system available.

### Package installation

Avoid installing packages wherever possible. Every package that is installed
must have an inline comment explaining why it is needed. Each package goes
on its own line to allow per-package comments:

```dockerfile
# Good
RUN apk add --no-cache \
    ca-certificates \
    # required for timezone handling in the scheduler
    tzdata

# Bad — no comments, packages on one line
RUN apk add --no-cache ca-certificates tzdata curl
```

Do not run `apt-get update` or `apk update` without a specific reason. If
a package is unavailable without an index update, document why in a comment.

### COPY and ADD

Always use `COPY` to copy files into the image. Never use `ADD` unless you
specifically need one of its unique features:

- Extracting a local tar archive automatically
- Fetching a file from a URL

When `ADD` is used, add a comment explaining why `COPY` is insufficient:

```dockerfile
# Good
COPY config/ /app/config/
COPY --from=builder /app/bin/mytool /mytool

# Only acceptable use of ADD
ADD archive.tar.gz /app/  # extracting tar — COPY does not support this
```

### ENTRYPOINT and CMD

Always use exec form — never shell form. Shell form spawns a shell process
as PID 1 which does not forward signals correctly:

```dockerfile
# Good — exec form
ENTRYPOINT ["/usr/local/bin/myapp"]
CMD ["--help"]

# Bad — shell form
ENTRYPOINT /usr/local/bin/myapp
CMD --help
```

Use `ENTRYPOINT` for the main executable. Use `CMD` for default arguments
that should be overridable at runtime. Never use `CMD` alone for compiled
binaries or long-running services.

### Multi-stage builds

Multi-stage builds are mandatory for all compiled languages (C++, Go). The
builder stage compiles the binary. The runtime stage is minimal and contains
only what is needed to run it:

```dockerfile
# Stage 1: Build
FROM alpine:3.20 AS builder
RUN apk add --no-cache \
    cmake \
    make \
    g++
WORKDIR /build
COPY . .
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build

# Stage 2: Runtime
FROM scratch
COPY --from=builder /build/build/bin/mytool /mytool
ENTRYPOINT ["/mytool"]
```

No multi-stage rule applies to interpreted languages — use project judgement.

### Healthchecks

Every Dockerfile must define a `HEALTHCHECK`. Use these standard defaults
unless the project has a specific reason to deviate:

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD <check command>
```

Choose the check command appropriate to the service:

```dockerfile
# HTTP service
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# TCP port check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD nc -z localhost 5432 || exit 1

# Process check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD pgrep mytool || exit 1
```

### Comments

Stage heading comments are always required:

```dockerfile
# Stage 1: Build
# Stage 2: Runtime
```

All other comments are sparse. Only add a comment for non-obvious decisions
or unusual procedures. Never write comments that narrate what the instruction
does — the instruction is self-evident:

```dockerfile
# Good — explains a non-obvious decision
# curl is unavailable in scratch — using wget from the builder stage
COPY --from=builder /usr/bin/wget /usr/bin/wget

# Bad — narrates the obvious
# Copy the application binary
COPY --from=builder /app/bin/mytool /mytool
```

Package comments are the exception — every non-obvious package always gets
a comment regardless of this rule.

### .dockerignore

A `.dockerignore` file is required when the build context contains files
that must not enter the image — secrets, credentials, large generated
directories, or local configuration files. It is optional but encouraged
for all other service directories.

---

## Docker Compose

### Structure

Each service lives in its own directory containing a `Dockerfile` and,
when needed, a `.dockerignore`. The location of service directories depends
on the project type:

**Standalone docker or infrastructure project** — service directories at
the project root:

```
api/
  Dockerfile
postgres/
  Dockerfile
docker-compose.yml
```

**Monorepo** — service directories under a `docker/` folder:

```
docker/
  api/
    Dockerfile
  postgres/
    Dockerfile
src/
docker-compose.yml
```

`docker-compose.yml` always lives at the project root.

### Build context

Every service must use a Dockerfile with an explicit build context. Never
use the `image:` key directly — even for unmodified third-party images.
This is a hard rule:

```yaml
# Good — always use a Dockerfile
services:
  postgres:
    build:
      context: ./postgres
      dockerfile: Dockerfile

# Bad — never reference an image directly
services:
  postgres:
    image: postgres:16.2-alpine
```

A Dockerfile for an unmodified third-party image contains only the `FROM`
line until customisation is needed:

```dockerfile
FROM postgres:16.2-alpine
```

### Version field

Never include the `version:` field. It is deprecated in Docker Compose V2
and must not be added:

```yaml
# Good
services:
  api:
    build:
      context: ./api
      dockerfile: Dockerfile

# Bad — version field is deprecated
version: "3.8"
services:
  api:
    ...
```

### Volumes

Use named volumes for all persistent data. Never use anonymous volumes —
they are untrackable and difficult to manage. Always declare named volumes
explicitly at the bottom of `docker-compose.yml`:

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

### Networking

Single-service projects do not need explicit network configuration — Docker
Compose provides a default network automatically.

Multi-service projects must define a named `backend` network for private
inter-service communication. Never expose internal services directly to the
host network. Always declare networks explicitly at the bottom of
`docker-compose.yml` alongside volumes:

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
