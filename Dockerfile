FROM golang:1.8

COPY . /go/src/ms-sendgrid-webhook
WORKDIR /go/src/ms-sendgrid-webhook
EXPOSE 3001
RUN make get
RUN make binary
CMD ["ms-sendgrid-webhook"]
