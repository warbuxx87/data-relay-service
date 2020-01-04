FROM alpine:3.9
RUN apk update && apk add --no-cache ca-certificates
RUN mkdir /app
COPY . /app
WORKDIR /app
EXPOSE 7070 7771
ENTRYPOINT [ "/app/cmd/data-relay-service/data-relay-service", "--prod"]
