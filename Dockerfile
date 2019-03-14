FROM golang AS builder

COPY . ${GOPATH}/src/github.com/ssoor/socks

RUN CGO_ENABLED=0 go build -o /socksd github.com/ssoor/socks/cmd/socksd && chmod +x /socksd

# Runtime

FROM scratch

COPY cmd/socksd/socksd.json /etc/socks/
# 将编译结果拷贝到容器中
COPY --from=builder /socksd /socks/socksd

ENTRYPOINT ["/socks/socksd"]

CMD [ "--config=/etc/socks/socksd.json" ]