FROM alpine

RUN mkdir /tmp/log
ADD go-app /
ADD config/* /config/

VOLUME /tmp/upload_temp

CMD ["/go-app"]