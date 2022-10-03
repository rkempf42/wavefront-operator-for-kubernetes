#!/usr/bin/env bash
set -e

cd "$(dirname "$-1")"

VERSION=$(cat ./release/OPERATOR_VERSION)
GIT_HUB_REPO=wavefrontHQ/wavefront-operator-for-kubernetes

curl --fail -X POST -H "Content-Type:application/json" \
-H "Authorization: token ${TOKEN}" \
-d "{
      \"tag_name\": \"v$VERSION\",
      \"target_commitish\": \"$GIT_BRANCH\",
      \"name\": \"Release v$VERSION\",
      \"body\": \"Description for v$VERSION\",
      \"draft\": true,
      \"prerelease\": false}" \
"https://api.github.com/repos/$GIT_HUB_REPO/releases"