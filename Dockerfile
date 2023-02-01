FROM golang:1.18-alpine AS build
WORKDIR /app
RUN apk add --update make
COPY . .
COPY config/config-example.yml config/config.yml
RUN make server-build

FROM golang:1.18-alpine AS deploy
WORKDIR /
COPY --from=build /app/bin/tcpserver /tcpserver
COPY --from=build /app/config/config.yml /config/config.yml
EXPOSE 8000
ENTRYPOINT [ "/tcpserver" ]
