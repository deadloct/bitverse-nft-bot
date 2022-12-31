FROM golang:1.19-alpine AS base
RUN apk update && apk add git gcc musl-dev
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o /bitverse-nft-bot

FROM golang:1.19-alpine
COPY --from=base /bitverse-nft-bot /bitverse-nft-bot
CMD ["/bitverse-nft-bot"]
