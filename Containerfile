FROM docker.io/library/golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
COPY webssh/go.mod webssh/go.sum ./webssh/
RUN go mod download && cd webssh && go mod download

COPY . .

# The browser terminal's SSH core, plus the wasm_exec.js shim from the same
# Go toolchain (they must match). Both are embedded into the site binary.
RUN cd webssh \
 && GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../static/ssh/killdeer-ssh.wasm . \
 && cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ../static/ssh/wasm_exec.js

RUN CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/killdeer-site .

FROM docker.io/library/alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/killdeer-site /usr/local/bin/killdeer-site

ENV PORT=8080

EXPOSE 8080

USER app

CMD ["killdeer-site"]
