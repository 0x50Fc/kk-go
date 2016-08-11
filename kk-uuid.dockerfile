FROM registry-internal.cn-hangzhou.aliyuncs.com/kk/kk-golang

RUN mkdir github.com

RUN mkdir github.com/hailongz

COPY . github.com/hailongz/kk-go

RUN go install github.com/hailongz/kk-go/kk-uuid

ENV KK_ADDR 127.0.0.1:87

CMD kk-uuid kk.uuid. $KK_ADDR
