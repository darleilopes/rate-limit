FROM golang:1.21 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o desafio

FROM scratch
WORKDIR /app
COPY --from=build /app/desafio .
COPY config.yaml .
ENTRYPOINT ["./desafio"]