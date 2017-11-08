FROM alpine:3.4

RUN apk update && apk add bash curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.7.5/bin/linux/amd64/kubectl && chmod +x kubectl && mv kubectl /usr/local/bin/

RUN mkdir -p /network /addons /root/.kube/ 
COPY network/*.yaml /network/
COPY addons/*.yaml /addons/
COPY entrypoint.sh /tmp/entrypoint.sh
ENTRYPOINT ["/tmp/entrypoint.sh"]
