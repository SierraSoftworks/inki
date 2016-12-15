FROM alpine

RUN apk add --update tini
ENTRYPOINT ["/sbin/tini", "--"]

ADD bin/inki /bin/inki

ENV PORT=3000
EXPOSE $PORT

ARG VERSION="development"
LABEL VERSION=$VERSION

WORKDIR /bin
CMD ["/bin/inki", "server", "-L", "INFO"]