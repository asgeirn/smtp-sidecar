#FROM golang AS builder
#WORKDIR /build
#ADD . .
#ENV CGO_ENABLED=0
#RUN go build

FROM scratch
EXPOSE 25
COPY smtp-sidecar /
ENV SMTP_LISTEN=:25
ENTRYPOINT [ "/smtp-sidecar" ]
