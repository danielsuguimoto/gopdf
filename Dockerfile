FROM golang:1.22-alpine

WORKDIR /usr/src/app
COPY . /usr/src/app

RUN apk add --no-cache \
      chromium \
      nss \
      freetype \
      freetype-dev \
      harfbuzz \
      ca-certificates \
      ttf-freefont \
      ghostscript

RUN go install github.com/air-verse/air@latest

RUN go mod tidy