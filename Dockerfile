FROM golang:latest AS builder

LABEL org.opencontainers.image.source https://github.com/MirageNetwork/MirageServer


RUN curl -o node.tar.gz https://nodejs.org/dist/v20.9.0/node-v20.9.0-linux-x64.tar.gz \
    && tar -xzf node.tar.gz -C /usr/local --strip-components=1 \
    && rm node.tar.gz
 
ENV PATH="/usr/local/bin:${PATH}"


WORKDIR /app

ADD MirageServer /app/MirageServer

RUN cd /app/MirageServer/cockpit_web && \
    npm install && npm run build && \
    cd /app/MirageServer/console_web && \
    npm install && npm run build && \
    go run /app/MirageServer/dist/build.go && \
    cd /app/MirageServer/dist/dist/ && \
    mv `basename ./*` ./mirageserver

FROM debian:stable
RUN apt-get update && \
    apt-get install -y openssl curl && \
    apt-get install -y libc6
WORKDIR /app
COPY --from=builder /app/MirageServer/dist/dist/mirageserver /app/mirageserver
RUN ["chmod", "+x", "/app/mirageserver"]
ENTRYPOINT [ "/app/mirageserver" ]
