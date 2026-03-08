# syntax=docker/dockerfile:1.20

############################
# Etapa de build ARM64
############################
FROM golang:1.26 AS builder
WORKDIR /src

ENV GOPROXY=https://proxy.golang.org,direct

# 1) Deps (capa estable + cache)
COPY go.mod go.sum ./
RUN go mod download

# 2) Codi
COPY . .

# 3) Build (cache de compilació)
RUN go build -o /out/myCalendar ./cmd/myCalendar

# --- runtime ---
FROM alpine:3.22
# copia el teu binari
COPY --from=builder /out/myCalendar /usr/local/bin/myCalendar

ENTRYPOINT ["/usr/local/bin/myCalendar"]
