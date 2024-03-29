FROM golang:latest as build

RUN mkdir -p /opt/lightheus
WORKDIR /opt/lightheus

COPY . .
RUN go build -ldflags "-extldflags '-static' -s -w"

# # update and install dependency
# RUN apk update && apk upgrade
# RUN apk add git

# RUN go get github.com/ppp225/lightheus
# RUN go build -ldflags "-extldflags '-static' -s -w" -o lightheus github.com/ppp225/lightheus

FROM justinribeiro/lighthouse

USER root

RUN mkdir -p /opt/lightheus
WORKDIR /opt/lightheus

COPY --from=build /opt/lightheus/lightheus /opt/lightheus/aetos-base.yml /opt/lightheus/lightheus.yml ./

# EXPOSE 22596

CMD ["./lightheus"]
