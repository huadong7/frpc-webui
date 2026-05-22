FROM alpine:3

RUN apk add --no-cache tzdata ca-certificates wget

COPY bin/frpc-linux-amd64 /usr/bin/frpc
RUN chmod +x /usr/bin/frpc

RUN mkdir -p /data
VOLUME ["/data"]

EXPOSE 7400

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:7400/healthz || exit 1

ENV FRPC_WEB_PORT=7400
ENV FRPC_WEB_USER=admin
ENV FRPC_WEB_PASSWORD=frpc@Admin1024
ENV FRPC_PROFILES_PATH=/data/frpc_profiles.json

ENTRYPOINT ["/usr/bin/frpc", \
  "--profiles-path", "/data/frpc_profiles.json", \
  "--web-addr", "0.0.0.0", \
  "--web-port", "7400"]
