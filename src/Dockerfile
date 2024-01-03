FROM golang:1.19 AS pxbuilder
COPY ./. /px
WORKDIR /px
RUN go build

FROM scratch
COPY --from=pxbuilder /px/px .