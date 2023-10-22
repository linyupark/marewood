# Build frontend dist.
FROM node:14-alpine AS frontend
WORKDIR /frontend-build
RUN npm config set registry https://registry.npm.taobao.org
COPY ./public .
RUN yarn && yarn build

# main ------------------------
FROM golang:1.21-alpine

# setting
WORKDIR /app
EXPOSE 8088
VOLUME  /app/resources

# installer
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk update
RUN apk add --no-cache build-base nodejs npm git wget
RUN npm config set registry https://registry.npm.taobao.org
RUN node -v && npm i -g pnpm && npm i -g yarn

# Install glibc (Bun dependancies)
RUN apk add gcompat
RUN wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub
RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.35-r0/glibc-2.35-r0.apk
RUN wget https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.35-r0/glibc-bin-2.35-r0.apk
RUN apk --no-cache --force-overwrite add glibc-2.35-r0.apk glibc-bin-2.35-r0.apk
RUN /usr/glibc-compat/bin/ldd /lib/ld-linux-x86-64.so.2
RUN npm i -g bun@1.0.2

COPY . .
COPY --from=frontend /frontend-build/build /app/public/build

# build go
RUN cd /app && CGO_ENABLED=1 go build -o MareWood ./MareWood.go

ENTRYPOINT ["/app/MareWood"]