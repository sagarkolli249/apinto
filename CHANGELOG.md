## [1.1.1](https://github.com/sagarkolli249/apinto/compare/v1.1.0...v1.1.1) (2026-02-18)


### Bug Fixes

* add helm dependency update step to CI pipeline ([bfd4e65](https://github.com/sagarkolli249/apinto/commit/bfd4e65a9ea09fe198b6554dd33684afc0b2161f))
* override container command to use start.sh for APINTO_DEBUG support ([57acd50](https://github.com/sagarkolli249/apinto/commit/57acd50760324f9e9cfcb720ff2739f6e450d441))

## [1.1.0](https://github.com/sagarkolli249/apinto/compare/v1.0.0...v1.1.0) (2026-02-18)


### Features

* **mistral:** add tool calling support with custom converter ([21cd8f5](https://github.com/sagarkolli249/apinto/commit/21cd8f543b4c96dce0877bb6be2ac236aaaf195a))


### Bug Fixes

* add APINTO_DEBUG env var to run in foreground mode ([d73ca33](https://github.com/sagarkolli249/apinto/commit/d73ca336ec34e230ca5d5c5ec418a726b8dcedee))
* add log volume mount for apinto startup ([29d798e](https://github.com/sagarkolli249/apinto/commit/29d798e87c482dbb281d17e95e455647ec708061))


### Documentation

* update deployment namespace to apipark ([1525351](https://github.com/sagarkolli249/apinto/commit/1525351a08a2ee6d5dcd91318bcabd04b5ec6bc2))

## 1.0.0 (2026-02-16)


### ⚠ BREAKING CHANGES

* Container now runs as UID 1000 instead of root
* Image tags changed from latest-amd64 to latest for amd64 architecture
* Image URLs changed from apinto-gateway to [secure]

Private: 865783518572.dkr.ecr.us-east-1.amazonaws.com/[secure]
Public:  public.ecr.aws/e5v3y2z9/apinto

### Features

* add container images section to README ([a33dc50](https://github.com/sagarkolli249/apinto/commit/a33dc5092d50e9b75a91fad853c37503971b1922))
* add deployment and verification scripts ([9e138ec](https://github.com/sagarkolli249/apinto/commit/9e138ec95c54d0f9a415cab6d4e7a6a3f02579a8))
* add non-root user support for PodSecurityContext ([5c14cd8](https://github.com/sagarkolli249/apinto/commit/5c14cd84756f24d4e0740a283dac3169838adbaf))
* convert apinto to StatefulSet with automated deployment ([917e5da](https://github.com/sagarkolli249/apinto/commit/917e5da75f4f782a2471e1ad465e4ed83cbcba9e))
* **k8s:** add Kubernetes deployment manifests and setup scripts ([2883e6f](https://github.com/sagarkolli249/apinto/commit/2883e6fceca448596519341a1e5c0d7dfe94c92f))
* migrate to existing ECR repositories ([47c88db](https://github.com/sagarkolli249/apinto/commit/47c88db3672a0031d274d4bf9da6fd846f836d63))
* set amd64 as default architecture without tag suffix ([18b6c06](https://github.com/sagarkolli249/apinto/commit/18b6c06f2349cc442707f707f013803b7b57f3ee))


### Bug Fixes

* add write permissions to workflow for automated version updates ([97b6532](https://github.com/sagarkolli249/apinto/commit/97b65325e84c9279bc32ecc3ac81bd0aa858d603))
* create volume directory before chown ([fd993ab](https://github.com/sagarkolli249/apinto/commit/fd993abe9d29a2f8078db3df070df7507c747b6d))


### Documentation

* add build and deployment verification summary ([cc528fc](https://github.com/sagarkolli249/apinto/commit/cc528fca307a1e181fbca139099ea86b09c770ec))
* add deployment configuration documentation ([3b5fbc1](https://github.com/sagarkolli249/apinto/commit/3b5fbc1df6a6aed0aae85118ba8d42d128390b41))
* add English documentation and highlight Bedrock enhancements ([6674182](https://github.com/sagarkolli249/apinto/commit/66741822d99f19f69adeec6afaea54297fe80d3a))
* add Kubernetes deployment quick start guide ([045bff2](https://github.com/sagarkolli249/apinto/commit/045bff2ec99ac22da00930bb1e4e8cacaf1a92e9))
