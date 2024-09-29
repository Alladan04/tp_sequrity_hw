FROM golang:1.21.0-alpine AS builder

COPY . /github.com/Alladan04/tp_security_hw
WORKDIR /github.com/Alladan04/tp_security_hw
RUN go mod download
RUN go clean --modcache
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -o ./.bin ./cmd/proxy/main.go

FROM scratch AS runner
WORKDIR /

COPY --from=builder /github.com/Alladan04/tp_security_hw/.bin .
COPY --from=builder /github.com/Alladan04/tp_security_hw/.env ./

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /
ENV TZ="Europe/Moscow"
ENV ZONEINFO=/zoneinfo.zip

EXPOSE 8080

ENTRYPOINT ["./.bin"]