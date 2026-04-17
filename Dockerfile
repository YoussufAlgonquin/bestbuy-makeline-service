# Stage 1: Build
# Go compiles to a single static binary — no runtime needed
FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o makeline-service .

# Stage 2: Run
# scratch is a completely empty image — just the binary. ~10MB total.
# This is the smallest possible Docker image.
FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
WORKDIR /app
COPY --from=build /app/makeline-service .
EXPOSE 8081
ENTRYPOINT ["./makeline-service"]