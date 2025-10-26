#!/bin/bash
export BUILDKIT_HOST=unix:///var/run/buildkit/buildkitd.sock
export APP_DIR="../railway-tests/rltest1"
export APP_REPO="https://github.com/laurci/rltest1.git"

export NEXT_PUBLIC_THING="abcd"

export GIT_AUTH_TOKEN=$(cat .env | grep GIT_AUTH_TOKEN | cut -d '=' -f 2)``
export REPORT_BASE_URL=$(cat .env | grep REPORT_BASE_URL | cut -d '=' -f 2)
export REPORT_AUTH_TOKEN=$(cat .env | grep REPORT_AUTH_TOKEN | cut -d '=' -f 2)

echo "GIT_AUTH_TOKEN: $GIT_AUTH_TOKEN"
echo "REPORT_BASE_URL: $REPORT_BASE_URL"
echo "REPORT_AUTH_TOKEN: $REPORT_AUTH_TOKEN"

buildctl build \
  --opt context=https://github.com/laurci/rltest1.git \
  --frontend=gateway.v0 \
  --opt source=ghcr.io/lttle-cloud/buildkit-frontend:latest \
  --opt report-build-id=1234567890 \
  --opt report-base-url=$REPORT_BASE_URL \
  --opt report-auth-token=$REPORT_AUTH_TOKEN \
  --opt github-token=$GIT_AUTH_TOKEN \
  --opt env-names=NEXT_PUBLIC_THING \
  --secret id=NEXT_PUBLIC_THING,env=NEXT_PUBLIC_THING \
  --secret id=GIT_AUTH_TOKEN,env=GIT_AUTH_TOKEN \
  --output type=image,name=ghcr.io/laurci/lttle-test:latest

echo "Successfully built image from git repo"

# buildctl build \
#   --local context=$APP_DIR \
#   --local dockerfile=$APP_DIR \
#   --frontend=gateway.v0 \
#   --opt source=ghcr.io/lttle-cloud/railpack-frontend:latest \
#   --output type=image,name=ghcr.io/laurci/lttle-test:latest

# echo "Successfully built image from local context and dockerfile"