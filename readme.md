# Readme

Query your Kubernetes cluster for images and check if they're outdated. Remember your developers to rebuild and redeploy their applications (based on your periodically updated base images) if the build timestamps of the currently deployed container images are older than a specified amount of time.

## Usage

    ./k8s-find-outdated-images -v find --context kind-kind --age 0h --filter type=workload --email email -o result.json

## Development

    go build
    go test
    go run . -v find --context kind-kind --age 3d --filter type=workload --email email