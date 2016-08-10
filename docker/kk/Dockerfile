FROM kk/golang

RUN go get -u github.com/hailongz/kk-go

RUN go install github.com/hailongz/kk-go

RUN go install github.com/hailongz/kk-go/kk-httpd

RUN go install github.com/hailongz/kk-go/kk-uuid

RUN mkdir /var/log/kk-go

EXPOSE 87
EXPOSE 80

CMD kk-go kk. --local 0.0.0.0:87 >> /var/log/kk-go/kk.log 2>>/var/log/kk-go/kk.log&

CMD kk-http kk.httpd. 127.0.0.1:87 :80 /kk/ >> /var/log/kk-go/kk-httpd.log 2>>/var/log/kk-go/kk-httpd.log&

CMD kk-uuid kk.uuid. 127.0.0.1:87 >> /var/log/kk-go/kk-uuid.log 2>>/var/log/kk-go/kk-uuid.log&

