FROM alpine

ARG PORT=6593
ENV ADDR=":${PORT}"

RUN apk --no-cache add sqlite-libs ca-certificates

COPY realm-linux-amd64 /realm
#COPY --from=0 /go/src/gitlab.brickchain.com/brickchain/realm/realm /realm
#COPY --from=0 /go/src/gitlab.brickchain.com/brickchain/realm/realm-migration/realm-migration /realm-migration
#COPY dev.yml /dev.yml

EXPOSE ${PORT}

ENTRYPOINT ["/realm"]
