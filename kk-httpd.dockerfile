FROM registry-internal.cn-hangzhou.aliyuncs.com/kk/golang

RUN mkdir github.com

RUN mkdir github.com/hailongz

COPY . github.com/hailongz/kk-go

RUN go install github.com/hailongz/kk-go/kk-httpd

ENV KK_ADDR 127.0.0.1:87

EXPOSE 88

CMD kk-httpd kk.httpd. $KK_ADDR :88
