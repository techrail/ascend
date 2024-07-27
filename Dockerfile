FROM golang:1.22-alpine AS builder

ARG GIT_URL
ARG BUILD_CMD="go build -tags netgo -ldflags '-s -w' -o app"

ENV GIT_URL=${GIT_URL}
ENV BUILD_CMD=${BUILD_CMD}

RUN apk add --no-cache git

WORKDIR /app

RUN git clone ${GIT_URL} .

RUN echo "#!/bin/sh" >> build.sh && echo -n "${BUILD_CMD}" >> build.sh && chmod +x build.sh
RUN ./build.sh

###########################<<<<<<<<<<>>>>>>>>>>>###########################

FROM alpine:latest
ARG START_CMD="./app"
ENV START_CMD=${START_CMD} 
ARG PORT="8080"
ENV PORT=${PORT}

WORKDIR /root/

COPY --from=builder /app/ .

EXPOSE ${PORT}

RUN echo "#!/bin/sh" >> run.sh && echo -n "${START_CMD}" >> run.sh && chmod +x run.sh
RUN chmod +x run.sh

# ENTRYPOINT ["tail"]
# CMD [ "-f", "/dev/null"]
ENTRYPOINT ["./run.sh"]