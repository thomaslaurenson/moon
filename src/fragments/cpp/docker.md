# C++ Docker

Supplements the Docker fragment. Applies to any tier that ships a binary: an application, or a library with a bundled CLI. A plain library ships no image and has no Dockerfile. Assumes the CMake fragment for the build directory and the testing option the recipe turns off.

## What it inherits

The Docker fragment's rules apply unchanged: a minor-version pin on the base image, official images only, `COPY` over `ADD`, exec-form `ENTRYPOINT`, stage heading comments and no others that narrate, OCI labels on the final stage, and the registry rules for tags, building once and publishing in a separate job. Two of its rules take their `scratch` form here: `USER` is a numeric id, since there is no passwd file to name one, and there is no `HEALTHCHECK`, because the image runs a command line tool rather than a service.

What this fragment settles is the recipe those rules leave open: one Dockerfile, named `Dockerfile`, building a statically linked musl binary into a `scratch` image.

## The Dockerfile

```dockerfile
# Stage 1: Build
FROM alpine:3.20 AS build

RUN apk add --no-cache \
    # The C and C++ toolchain, with the musl-dev the static link needs
    build-base \
    cmake

WORKDIR /src
COPY . .

# -static is what makes the scratch stage viable: the binary carries musl and
# libstdc++ with it and needs no loader at runtime.
RUN cmake -B build/release \
        -DCMAKE_BUILD_TYPE=Release \
        -DMYPROJ_BUILD_TESTING=OFF \
        -DCMAKE_EXE_LINKER_FLAGS="-static" \
    && cmake --build build/release --parallel $(nproc) \
    && strip build/release/bin/myproj

# Stage 2: Runtime
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/<owner>/<repo>"
LABEL org.opencontainers.image.description="One sentence, the same as the repository description"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=build /src/build/release/bin/myproj /myproj
USER 1001:1001
ENTRYPOINT ["/myproj"]
```

- `alpine` is the builder because its libc is musl, and `-static` against musl is what produces a binary with no loader and no libc version floor; see One Linux binary, not two in cpp/workflows-app.md for why that is the Linux artefact. The Docker fragment prefers Alpine anyway, so nothing is overridden. Do not copy the tag from this document as the version to match: Dependabot keeps it current (see the Dependabot fragment).
- `-DCMAKE_EXE_LINKER_FLAGS="-static"` is not optional, and leaving it out fails in a way that is easy to miss: the image builds fine, and the container then exits immediately with `no such file or directory` on a binary that plainly exists. What is missing is `/lib/ld-musl-x86_64.so.1`, the dynamic loader, which `scratch` does not have. The smoke run in `build.yml` is what catches it in CI (see cpp/workflows-app.md); locally, `docker run --rm <image> --version`, or `file` on the extracted binary, which should say `statically linked`.
- `MYPROJ_BUILD_TESTING=OFF` keeps Catch2 out of an image that never runs tests, and `strip` cuts the binary substantially.
- The build goes into `build/release`, the directory cpp/cmake.md reserves for the shipped artifact: optimised, testing off, nothing runs `ctest` against it.
- `build-base` rather than the individual packages, because it is Alpine's own name for the C and C++ toolchain and pulls in `musl-dev`, which the static link needs and which is easy to leave out of a hand-written list.
- The whole checkout is the build context, submodules included: the image builds from the same `extern/` the tests do, so `docker build` follows `git submodule update --init` locally and a checkout with `submodules: true` in CI.

## .dockerignore

Required, where the Docker fragment merely encourages it, because the context otherwise carries every local build directory into the image build:

```text
build/
.git/
test/data/
```

`build/` is the point: a developer's `build/dev`, `build/lint` and the rest are large, stale and irrelevant to the image. `.git/` is history the build never reads. `test/data/` may hold integration inputs that are gitignored precisely because they are large. Never ignore `extern/`; the submodules are the build.

## One Dockerfile

The file is `Dockerfile`, with no suffix, so `docker build .` finds it and nothing needs `-f`. There is one because there is one Linux artifact: a static musl binary runs on every distribution, where a glibc build from the builder's distribution would not, and a second Dockerfile would be a strictly narrower duplicate of the first. The reasoning is in cpp/workflows-app.md.

## Publishing

`build.yml` builds the image and saves it as a tar; `release.yml` and `prerelease.yml` load and push those bytes. The jobs are in cpp/release-app.md, and they follow the Docker fragment's registry rules to the letter: `<version>` and `latest` on a release, `dev` on the rolling prerelease and never `latest`, and the push in a job of its own gated on the release having succeeded.
