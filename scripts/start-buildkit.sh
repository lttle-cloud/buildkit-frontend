#!/bin/bash

mkdir -p /var/run/buildkit
docker run -d --name buildkitd --privileged \
  -v /var/run/buildkit:/run/buildkit \
  -v buildkit-data:/var/lib/buildkit \
  moby/buildkit:latest \
  --addr unix:///run/buildkit/buildkitd.sock
