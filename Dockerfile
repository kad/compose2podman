FROM scratch

COPY compose2podman /usr/local/bin/compose2podman

ENTRYPOINT ["/usr/local/bin/compose2podman"]
CMD ["-version"]
