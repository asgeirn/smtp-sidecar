#FROM golang AS builder
#WORKDIR /build
#ADD . .
#ENV CGO_ENABLED=0
#RUN go build

FROM scratch
USER 1000
EXPOSE 2525
COPY smtp-sidecar /
ENV SMTP_LISTEN=:2525
ENTRYPOINT [ "/smtp-sidecar" ]
