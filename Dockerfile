FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# A static binary, so the runtime image needs no C library. The SQLite driver is
# pure Go, so cgo is not needed either.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/velo ./cmd/velo

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=build /out/velo /usr/local/bin/velo
COPY data ./data

ENV DB_DATABASE=/app/database/database.sqlite \
    RIDES_JSON_PATH=/app/data/rides.json \
    HTTP_PORT=8000

EXPOSE 8000

ENTRYPOINT ["velo"]
CMD ["serve"]
