FROM golang:1.14

COPY . /code

WORKDIR /code

RUN CGO_ENABLED=0 go build .

FROM scratch
COPY --from=0 /code/consensus-backend /consensus-backend
ENTRYPOINT ["/consensus-backend"]
