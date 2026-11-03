# healthprobe

A tiny, dependency-free Go library and CLI that runs one HTTP or TCP health check and exits 0 (healthy) or 1 (unhealthy). It exists for one job: Docker `HEALTHCHECK` instructions in images that have no shell, `curl` or `wget`, such as distroless and `scratch` images.

## Install

```
go install github.com/PrinceAdvant/healthprobe/cmd/healthprobe@latest
```

## Use it in a Dockerfile

Build it in the same builder stage as your app and copy the static binary into the final image:

```dockerfile
FROM golang:1.23 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/app
RUN CGO_ENABLED=0 go install -ldflags="-s -w" github.com/PrinceAdvant/healthprobe/cmd/healthprobe@latest

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/app /app
COPY --from=build /go/bin/healthprobe /healthprobe
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
  CMD ["/healthprobe", "-q", "http://127.0.0.1:8080/healthz"]
ENTRYPOINT ["/app"]
```

Use the exec form (`CMD ["/healthprobe", ...]`). The shell form needs `/bin/sh`, which distroless images do not have. The stripped binary adds roughly 5 to 6 MB to the image.

## CLI

```
healthprobe http://127.0.0.1:8080/healthz
healthprobe -contains '"status":"ok"' http://127.0.0.1:8080/healthz
healthprobe -H 'Host: app.internal' http://127.0.0.1/healthz
healthprobe tcp://127.0.0.1:5432
```

- `-timeout` limits the check (default 3s). Keep it below the `HEALTHCHECK --timeout`.
- `-min-status` and `-max-status` set the accepted status range (default 200 to 399). Redirects are not followed.
- `-contains` requires text in the first 64 KiB of the body.
- `-q` prints nothing on success, which keeps `docker inspect` health logs short.

## Library

```go
err := healthprobe.Check(ctx, "http://127.0.0.1:8080/healthz", healthprobe.Options{Timeout: 2 * time.Second})
```

## Health checks that help

A health check is only as good as the endpoint behind it. Check that the process can serve requests, and be careful about including external dependencies: if the database is down, marking every app container unhealthy can trigger restarts that make recovery slower. In Compose, combine health checks with `depends_on: condition: service_healthy` so services start in the right order. More patterns and pitfalls are in this guide to [Docker Compose healthchecks](https://peon.sh/blogs/docker-compose-healthchecks).

Maintained by people who work on [Peon](https://peon.sh/), an open-source self-hosted deployment platform. MIT licensed.
