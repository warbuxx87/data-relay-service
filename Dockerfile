FROM golang:alpine3.10
RUN mkdir /app
COPY . /app
WORKDIR /app
# install gcc and other dependencies
# gcc is required by sqlite database that is required to run this app
RUN set -ex && apk add --no-cache gcc musl-dev git
RUN apk update && apk add --no-cache ca-certificates
# download deps and build app
RUN go get ./...
RUN go build -o data-relay-service cmd/data-relay-service/main.go
EXPOSE 7070
EXPOSE 7771
ENTRYPOINT [ "/app/data-relay-service", "-prod"]