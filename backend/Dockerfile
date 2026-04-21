FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/backend ./cmd/server

FROM alpine:3.22

RUN adduser -D -H backend
USER backend
WORKDIR /app

COPY --from=build /out/backend /app/backend

EXPOSE 8080
ENTRYPOINT ["/app/backend"]
