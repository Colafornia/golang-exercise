FROM golang:1.13.3

ENV GOPROXY="https://goproxy.io"

ENV TZ=Asia/Shanghai

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN mkdir /app

ADD . /app/

WORKDIR /app

COPY . .

RUN cd fund-valuation-monitor && go build -o main .

CMD ["/app/fund-valuation-monitor/main"]
