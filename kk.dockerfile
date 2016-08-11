FROM registry-internal.cn-hangzhou.aliyuncs.com/kk/kk-golang

RUN mkdir github.com

RUN mkdir github.com/hailongz

COPY . github.com/hailongz/kk-go

RUN go install github.com/hailongz/kk-go

EXPOSE 87

CMD kk-go kk. --local 0.0.0.0:87
