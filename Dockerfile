FROM    golang:1.17.3-alpine3.13 AS stage1
RUN     apk update && apk add git
ARG     GOPRIVATE=git.zbdigital.net
ARG     GIT_USER=zb02
ARG     GIT_TOKEN=9ECHKy7z8jXdSN8WdnJR
RUN     git config --global url."https://$GIT_USER:$GIT_TOKEN@$GOPRIVATE".insteadOf "https://$GOPRIVATE"
ENV     RUN_PATH=/app PROJ_PATH=/build
RUN     mkdir -p $RUN_PATH
WORKDIR $RUN_PATH
ENV     GO111MODULE=on
ENV     GOPRIVATE=git.zbdigital.net/*
COPY    go.mod .
COPY    go.sum .
RUN     go mod download

FROM    stage1 AS stage2
RUN     apk update && apk add make
USER    root
ADD     . $PROJ_PATH
WORKDIR $PROJ_PATH
RUN     make build pack unpack path=$RUN_PATH

FROM    alpine
USER    root
ENV     RUN_PATH=/app
RUN     mkdir -p $RUN_PATH
RUN     apk add tzdata && cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && echo "Asia/Taipei" >  /etc/timezone
RUN     date
RUN     apk del tzdata
COPY    --from=stage2 ${RUN_PATH} ${RUN_PATH}
WORKDIR ${RUN_PATH}
EXPOSE  8080
ENTRYPOINT ["./app"]