FROM docker.io/library/golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/killdeer-site .

FROM docker.io/library/alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/killdeer-site /usr/local/bin/killdeer-site

ENV PORT=8080

EXPOSE 8080

USER app

CMD ["killdeer-site"]
