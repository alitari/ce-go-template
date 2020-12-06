FROM golang:1.14.3-alpine AS builder
ARG main_path
RUN echo $MAIN_PATH
WORKDIR /work
COPY . /work
RUN go build -o bin/producer $main_path

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /work/bin/producer .
CMD ["./producer"]