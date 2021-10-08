#!/bin/bash
echo "BUILD_SCM_VERSION $(git rev-parse --short=7 HEAD)"
echo "BUILD_CURRENT_BRANCH $(git rev-parse --abbrev-ref HEAD)"