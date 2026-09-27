# syntax=docker/dockerfile:1
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/server .

# Distroless: no shell or package manager, and it runs as a non-root user.
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/server /server
EXPOSE 8080
# Numeric UID of distroless "nonroot": with runAsNonRoot the kubelet refuses a user *name*,
# because it cannot prove the name does not map to UID 0.
USER 65532:65532
ENTRYPOINT ["/server"]
