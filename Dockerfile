FROM golang:1.20-alpine as deploy-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -buildvcs=false -trimpath  -o app 

FROM debian:bullseye-slim as deploy

RUN apt-get update

COPY --from=deploy-builder /app/app .

CMD ["./app"]

# Goでホットリロード環境を実現する, airコマンドを実行するとファイルの変更を検知するたびgo buildをする
FROM golang:1.20-alpine as dev
WORKDIR /app
RUN go install github.com/cosmtrek/air@latest
CMD ["air"]