FROM alpine:latest
COPY gipgee /bin/gipgee
ENTRYPOINT ["/bin/gipgee"]
