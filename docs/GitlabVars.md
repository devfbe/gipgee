# Gitlab env vars

This document describes the values of the different gitlab env vars depending on the action a user performed and on the state of the repository.

It's a necessary documentation for the image build pipeline generator that has to decide which images need to be rebuilt depending on the repository state.

Below, there are the different use cases showing the relevant env vars (dumped with env).

The env vars described have been tested with Gitlab 15.2.2.
- [Gitlab env vars](#gitlab-env-vars)
  - [Initial commit after push](#initial-commit-after-push)
  - [Initial commit - manual trigger](#initial-commit---manual-trigger)
  - [Normal non initial push on default branch](#normal-non-initial-push-on-default-branch)
  - [Manual trigger on repo with multiple commits](#manual-trigger-on-repo-with-multiple-commits)
  - [Merged merge request](#merged-merge-request)
  - [Feature branch push (initial commit in feature branch)](#feature-branch-push-initial-commit-in-feature-branch)
  - [Feature branch push (non initial commit in feature branch)](#feature-branch-push-non-initial-commit-in-feature-branch)
  - [Feature branch force push](#feature-branch-force-push)

## Initial commit after push

```sh
CI_COMMIT_BRANCH=main
CI_BUILD_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_REF_NAME=main
CI_BUILD_REF_NAME=main
CI_COMMIT_REF_SLUG=main
CI_PIPELINE_SOURCE=push
CI_COMMIT_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_SHA=1bf1b017840970833f848165f8d1e67e601b2569
CI_BUILD_REF_SLUG=main
```

## Initial commit - manual trigger

```sh
CI_COMMIT_BRANCH=main
CI_BUILD_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_REF_NAME=main
CI_BUILD_REF_NAME=main
CI_COMMIT_REF_SLUG=main
CI_PIPELINE_SOURCE=web
CI_COMMIT_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_SHA=1bf1b017840970833f848165f8d1e67e601b2569
CI_BUILD_REF_SLUG=main
```

## Normal non initial push on default branch

```sh
CI_COMMIT_BRANCH=main
CI_BUILD_BEFORE_SHA=1bf1b017840970833f848165f8d1e67e601b2569
CI_COMMIT_REF_NAME=main
CI_BUILD_REF_NAME=main
CI_COMMIT_REF_SLUG=main
CI_PIPELINE_SOURCE=push
CI_COMMIT_BEFORE_SHA=1bf1b017840970833f848165f8d1e67e601b2569
CI_COMMIT_SHA=24301d9462995e670c11626c5912036ae14c183d
CI_BUILD_REF_SLUG=main
```

## Manual trigger on repo with multiple commits

```sh
CI_COMMIT_BRANCH=main
CI_BUILD_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_REF_NAME=main
CI_BUILD_REF_NAME=main
CI_COMMIT_REF_SLUG=main
CI_PIPELINE_SOURCE=web
CI_COMMIT_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_SHA=24301d9462995e670c11626c5912036ae14c183d
CI_BUILD_REF_SLUG=main
```

## Merged merge request

```sh
CI_COMMIT_BRANCH=main
CI_BUILD_BEFORE_SHA=24301d9462995e670c11626c5912036ae14c183d
CI_COMMIT_REF_NAME=main
CI_BUILD_REF_NAME=main
CI_COMMIT_REF_SLUG=main
CI_PIPELINE_SOURCE=push
CI_COMMIT_BEFORE_SHA=24301d9462995e670c11626c5912036ae14c183d
CI_COMMIT_SHA=96ea3a1d5f1872946748ba10ff147cf83cfb2f57
CI_BUILD_REF_SLUG=main
```

## Feature branch push (initial commit in feature branch)

```sh
CI_COMMIT_BRANCH=testbranch
CI_BUILD_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_REF_NAME=testbranch
CI_BUILD_REF_NAME=testbranch
CI_COMMIT_REF_SLUG=testbranch
CI_PIPELINE_SOURCE=push
CI_COMMIT_BEFORE_SHA=0000000000000000000000000000000000000000
CI_COMMIT_SHA=a175a95833598723d50e9f5f7896d793bd9e2a52
CI_BUILD_REF_SLUG=testbranch
```

## Feature branch push (non initial commit in feature branch)

```sh
CI_COMMIT_BRANCH=testbranch
CI_BUILD_BEFORE_SHA=a175a95833598723d50e9f5f7896d793bd9e2a52
CI_COMMIT_REF_NAME=testbranch
CI_BUILD_REF_NAME=testbranch
CI_COMMIT_REF_SLUG=testbranch
CI_PIPELINE_SOURCE=push
CI_COMMIT_BEFORE_SHA=a175a95833598723d50e9f5f7896d793bd9e2a52
CI_COMMIT_SHA=9a76a0d88d1f999a0a740952739474f54cc422b5
CI_BUILD_REF_SLUG=testbranch
```

## Feature branch force push

```sh
CI_COMMIT_BRANCH=testbranch
CI_BUILD_BEFORE_SHA=28ee30c9ef3a91dcc06d0bdf04ec53b76eadfe92
CI_COMMIT_REF_NAME=testbranch
CI_BUILD_REF_NAME=testbranch
CI_COMMIT_REF_SLUG=testbranch
CI_PIPELINE_SOURCE=push
CI_COMMIT_BEFORE_SHA=28ee30c9ef3a91dcc06d0bdf04ec53b76eadfe92
CI_COMMIT_SHA=9891a1d24de7049244837cef2fde3c6134b6ac4e
CI_BUILD_REF_SLUG=testbranch
```
