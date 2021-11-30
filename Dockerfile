FROM golang:1.17 AS builder

WORKDIR /opt

COPY . .

RUN go build

FROM public.ecr.aws/lambda/provided:al2 AS runtime

COPY --from=builder /opt/jubilant-happiness /usr/local/bin/jubilant-happiness

ENTRYPOINT ["/usr/local/bin/jubilant-happiness"]
