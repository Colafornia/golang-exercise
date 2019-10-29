FROM golang:alpine

ENV GOLANGCI_LINT_VERSION=1.16.0

ENV GOPROXY="https://goproxy.io"

ENV EMAIL_NAME=$value1 EMAIL_PASSWORD=$value2

RUN echo "Asia/Shanghai" > /etc/timezone

RUN mkdir /app

ADD . /app/

WORKDIR /app

COPY . .

RUN cd fund-valueation-monitor && go build -o main .

CMD ["/app/fund-valueation-monitor/main"]
