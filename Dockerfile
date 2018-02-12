FROM debian:stretch

WORKDIR /app

# copy binary into image
COPY git2consul /app/git2consul

CMD ["/app/git2consul"]
