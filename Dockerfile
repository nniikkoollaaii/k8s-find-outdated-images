FROM scratch
ENTRYPOINT ["/k8s-find-outdated-images"]
COPY k8s-find-outdated-images /