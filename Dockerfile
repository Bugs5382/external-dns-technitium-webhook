FROM gcr.io/distroless/static-debian12:nonroot

USER 20000:20000
COPY --chmod=555 external-dns-technitium-webhook /opt/external-dns-technitium-webhook/app

ENTRYPOINT ["/opt/external-dns-technitium-webhook/app"]