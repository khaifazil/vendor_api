#1st stage docker build
#get the image and the version. If image is in local it will get from local, else will download from repo
#AS = give alias
FROM golang:1.18.3-alpine3.16 AS builder
#set the working directory
WORKDIR /app
#copy all from current directory to working directory. first . means all files, second . is the destination with . as context.
COPY . .
#run in container's terminal this command.
#-o means output as <nameofbinary>
RUN go build -o main

#2nd stage docker build
#get ubuntu version. best to be same version as image in 1st stage
FROM alpine:3.16
#set the working directory in 2nd build
WORKDIR /app
#copy from builder stage - path - destination(working directory)
COPY --from=builder /app/main .
#copy <directory to copy from host> ./<new Directory in 2nd stage container>
COPY logs ./logs
COPY SSL ./SSL
COPY app.env .
COPY wait-for.sh .

#expose the specified port for later use when running the container
EXPOSE 9091
#run main binary after image is built
#CMD happens only at the end and there can only be one
CMD ["/app/main"]