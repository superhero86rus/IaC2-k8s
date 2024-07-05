FROM golang as builder

ARG CGO_ENABLED=0
ARG GOOS=linux

# RUN apt update && apt install -y upx

WORKDIR /build
COPY . .
RUN test -e go.mod || go mod init gowebd

# удаляем отладочную информацию из бинарника
RUN go build -ldflags "-s -w" -o /gowebd 
# сжимаем бинарник
# RUN time upx --best --lzma /gowebd

# можно еще scratch
FROM alpine
COPY --from=builder /gowebd /gowebd

ENTRYPOINT ["/gowebd"]