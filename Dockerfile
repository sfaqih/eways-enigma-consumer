####### golang alpine 1.13.5 ########
# FROM golang@sha256:0991060a1447cf648bab7f6bb60335d1243930e38420bee8fec3db1267b84cfa as builder
FROM golang:1.19.3-alpine as builder

ENV GO111MODULE=on
# ENV GOFLAGS=-mod=vendor
ENV GIT_TERMINAL_PROMPT=1


WORKDIR '/app'

COPY go.mod .
COPY go.sum .

# install git inside container in order to run go get command
RUN apk update && apk add git

RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta

RUN echo -e "machine gitlab.com\nlogin devwgroup\npassword glpat-WxV7Fak1msMzWQSBKofC" > ~/.netrc
RUN git config --global user.name "devwgroup"
RUN git config --global user.password "Qe*E46c3h6$I"
RUN git config --global user.email "dev.wgroup@gmail.com"

COPY . .
RUN git config credential.helper store
RUN go env -w GOPRIVATE=github.com/AplikasiRentasDigital/eways-enigma-master
RUN go get -d github.com/AplikasiRentasDigital/eways-enigma-master@superenigma
# RUN git remote set-url origin https://wgroup249:a99i9Y5Gz665zXiRaWLZ@gitlab.com/wgroup1/luna-consumer
# RUN go get ssh://wgroup249:76ZNtQE/z+%bt@gitlab.com/wgroup1/luna
RUN go mod tidy
RUN go mod download
RUN go mod vendor
RUN go mod verify

RUN go build -o enigmaconsumer

CMD ["./enigmaconsumer"]