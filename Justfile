# version tag

VERSION := "latest"
REGISTRY_DEV := "europe-docker.pkg.dev/azin-dev/builder"

build-frontend-image:
    docker build --platform linux/amd64 --provenance=false -f Dockerfile.frontend -t {{ REGISTRY_DEV }}/buildkit-frontend:{{ VERSION }} . --push

build-analyzer:
    go build -o ./target/analyzer analyzer/main.go

build-analyzer-image:
    docker build --platform linux/amd64 --provenance=false -f Dockerfile.analyzer -t {{ REGISTRY_DEV }}/buildkit-analyzer:{{ VERSION }} . --push

build-images: build-frontend-image build-analyzer-image
