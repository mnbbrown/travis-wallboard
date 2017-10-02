FROM golang:alpine
RUN apk add ca-certificates
RUN mkdir /app
ADD __build__/travis-wallboard /app/travis-wallboard
# Set the working directory to the app directory
WORKDIR /app
EXPOSE 8000
ENTRYPOINT /app/travis-wallboard
