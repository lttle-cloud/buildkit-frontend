# version tag
VERSION := "latest"

build-frontend-image:
    docker build -f Dockerfile.frontend -t ghcr.io/lttle-cloud/buildkit-frontend:{{VERSION}} . --push

build-analyzer:
    go build -o ./target/analyzer analyzer/main.go

build-analyzer-image:
    docker build -f Dockerfile.analyzer -t ghcr.io/lttle-cloud/buildkit-analyzer:{{VERSION}} . --push

build-images: build-frontend-image build-analyzer-image