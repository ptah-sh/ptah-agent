# Changelog

## [0.23.0](https://github.com/ptah-sh/ptah-agent/compare/v0.22.0...v0.23.0) (2024-10-14)


### Features

* [#60](https://github.com/ptah-sh/ptah-agent/issues/60) support removing files from s3 storages ([78ae011](https://github.com/ptah-sh/ptah-agent/commit/78ae01161edd840f90275750e9e17c9975b3c71e))
* [#61](https://github.com/ptah-sh/ptah-agent/issues/61) support replicatedjob-based backups ([7e8a165](https://github.com/ptah-sh/ptah-agent/commit/7e8a165e83f2a6325a141784972314565c8652e7))
* [#63](https://github.com/ptah-sh/ptah-agent/issues/63) support downloading files from s3 storage ([a0cd1e3](https://github.com/ptah-sh/ptah-agent/commit/a0cd1e370134d4f2048a8adba0e799af163efc2e))


### Bug Fixes

* [#61](https://github.com/ptah-sh/ptah-agent/issues/61) service start monitoring for new services ([ecb13d9](https://github.com/ptah-sh/ptah-agent/commit/ecb13d927f5c4f28025f37de51f3843f9b3b1175))

## [0.22.0](https://github.com/ptah-sh/ptah-agent/compare/v0.21.0...v0.22.0) (2024-10-05)


### Features

* [#58](https://github.com/ptah-sh/ptah-agent/issues/58) send caddy and system metrics to ptah-server ([7fed58c](https://github.com/ptah-sh/ptah-agent/commit/7fed58c41e4d7e2c88a0c0acfe629471a0299197))

## [0.21.0](https://github.com/ptah-sh/ptah-agent/compare/v0.20.0...v0.21.0) (2024-09-24)


### Features

* [#56](https://github.com/ptah-sh/ptah-agent/issues/56) support docker registry username encryption ([a5df51a](https://github.com/ptah-sh/ptah-agent/commit/a5df51a7832983a8665377860766e2b23acd9a07))

## [0.20.0](https://github.com/ptah-sh/ptah-agent/compare/v0.19.0...v0.20.0) (2024-09-22)


### Features

* [#54](https://github.com/ptah-sh/ptah-agent/issues/54) support secrets encryption ([367eab9](https://github.com/ptah-sh/ptah-agent/commit/367eab9adcf0f1ebf742e41c4b78909afc710bb0))

## [0.19.0](https://github.com/ptah-sh/ptah-agent/compare/v0.18.0...v0.19.0) (2024-09-16)


### Features

* [#52](https://github.com/ptah-sh/ptah-agent/issues/52) add the launch service task ([ae3b26f](https://github.com/ptah-sh/ptah-agent/commit/ae3b26f180f8e911e7dcc2b8acaf9f86aa6fc6f2))

## [0.18.0](https://github.com/ptah-sh/ptah-agent/compare/v0.17.2...v0.18.0) (2024-09-13)


### Features

* [#49](https://github.com/ptah-sh/ptah-agent/issues/49) drop the 'preserve secrets' feature ([11d3f5c](https://github.com/ptah-sh/ptah-agent/commit/11d3f5c86c0f62720f6efde1655908e2d6a22792))
* [#51](https://github.com/ptah-sh/ptah-agent/issues/51) drop secrets vars configs support ([688c6c6](https://github.com/ptah-sh/ptah-agent/commit/688c6c6076a7e4bda2a3824265d0c8c3bfe1be97))

## [0.17.2](https://github.com/ptah-sh/ptah-agent/compare/v0.17.1...v0.17.2) (2024-09-08)


### Bug Fixes

* always provide addr ([db08b7b](https://github.com/ptah-sh/ptah-agent/commit/db08b7b13b1b03b35f8fcc6e63b21b5331bbee89))

## [0.17.1](https://github.com/ptah-sh/ptah-agent/compare/v0.17.0...v0.17.1) (2024-09-08)


### Bug Fixes

* try set user for backups ([d62e305](https://github.com/ptah-sh/ptah-agent/commit/d62e3056b7f45b2a2185318a82c7894a60f49b48))

## [0.17.0](https://github.com/ptah-sh/ptah-agent/compare/v0.16.0...v0.17.0) (2024-09-01)


### Features

* [#44](https://github.com/ptah-sh/ptah-agent/issues/44) allow execute tasks from JSON file ([b935b01](https://github.com/ptah-sh/ptah-agent/commit/b935b01cbedd2e3f7b62b58a5572e6bac7313e76))

## [0.16.0](https://github.com/ptah-sh/ptah-agent/compare/v0.15.1...v0.16.0) (2024-08-31)


### Features

* [#44](https://github.com/ptah-sh/ptah-agent/issues/44) allow to list available ips ([214fa59](https://github.com/ptah-sh/ptah-agent/commit/214fa59dfbd43005f2a593582d950e4e772327c5))

## [0.15.1](https://github.com/ptah-sh/ptah-agent/compare/v0.15.0...v0.15.1) (2024-08-25)


### Bug Fixes

* [#39](https://github.com/ptah-sh/ptah-agent/issues/39) restart dird ([848bd1b](https://github.com/ptah-sh/ptah-agent/commit/848bd1b0de3a3d064a4685e2bc68947d58b2a8c1))
* [#39](https://github.com/ptah-sh/ptah-agent/issues/39) restart dird ([902bcbe](https://github.com/ptah-sh/ptah-agent/commit/902bcbe610b66d252bc24ea7d1f13f39833c16df))

## [0.15.0](https://github.com/ptah-sh/ptah-agent/compare/v0.14.0...v0.15.0) (2024-08-25)


### Features

* [#39](https://github.com/ptah-sh/ptah-agent/issues/39) support dird (re-) configuration ([90c1772](https://github.com/ptah-sh/ptah-agent/commit/90c177240ea950f97fcc125841b6e721f47c6f97))

## [0.14.0](https://github.com/ptah-sh/ptah-agent/compare/v0.13.0...v0.14.0) (2024-08-25)


### Features

* [#37](https://github.com/ptah-sh/ptah-agent/issues/37) better address field name ([28763b4](https://github.com/ptah-sh/ptah-agent/commit/28763b4a737de492994a2395dfb554ba3020082a))

## [0.13.0](https://github.com/ptah-sh/ptah-agent/compare/v0.12.1...v0.13.0) (2024-08-25)


### Features

* [#37](https://github.com/ptah-sh/ptah-agent/issues/37) send the current node (internal) addr on startup ([01b68e9](https://github.com/ptah-sh/ptah-agent/commit/01b68e95bb97293cf70cb3d8aa3b44027e6bfd02))


### Bug Fixes

* [#34](https://github.com/ptah-sh/ptah-agent/issues/34) close the reader to avoid leaks ([bc29b35](https://github.com/ptah-sh/ptah-agent/commit/bc29b353c9e6635b1522326a6c94898b432f7ffe))

## [0.12.1](https://github.com/ptah-sh/ptah-agent/compare/v0.12.0...v0.12.1) (2024-08-20)


### Bug Fixes

* [#34](https://github.com/ptah-sh/ptah-agent/issues/34) pull d3fk/s3cmd image before starting a container ([d841653](https://github.com/ptah-sh/ptah-agent/commit/d8416530b1f4203851a6b7fd70664ca8a4aeaecf))

## [0.12.0](https://github.com/ptah-sh/ptah-agent/compare/v0.11.1...v0.12.0) (2024-08-09)


### Features

* [#31](https://github.com/ptah-sh/ptah-agent/issues/31) ignore docker network interfaces ([cbe7a08](https://github.com/ptah-sh/ptah-agent/commit/cbe7a08524ce5ab666d6f870960658a1d6a52674))

## [0.11.1](https://github.com/ptah-sh/ptah-agent/compare/v0.11.0...v0.11.1) (2024-08-09)


### Bug Fixes

* [#29](https://github.com/ptah-sh/ptah-agent/issues/29) mark failed to parse tasks as failed ([1c6bde9](https://github.com/ptah-sh/ptah-agent/commit/1c6bde96abf6c7d0263999d034eac1dc02e56f76))

## [0.11.0](https://github.com/ptah-sh/ptah-agent/compare/v0.10.0...v0.11.0) (2024-07-30)


### Features

* [#27](https://github.com/ptah-sh/ptah-agent/issues/27) handle joinswarm command ([3680760](https://github.com/ptah-sh/ptah-agent/commit/3680760d72d44a45c77258c939d225ff66503773))

## [0.10.0](https://github.com/ptah-sh/ptah-agent/compare/v0.9.0...v0.10.0) (2024-07-16)


### Features

* [#5](https://github.com/ptah-sh/ptah-agent/issues/5) test empty commit 2 ([61e6a11](https://github.com/ptah-sh/ptah-agent/commit/61e6a1122fb091339aa838975b13b661eb74cdaf))

## [0.9.0](https://github.com/ptah-sh/ptah-agent/compare/v0.8.0...v0.9.0) (2024-07-16)


### Features

* [#5](https://github.com/ptah-sh/ptah-agent/issues/5) test empty commit ([d0fb851](https://github.com/ptah-sh/ptah-agent/commit/d0fb851b5cbcf2042471afe47408cfae11d96af1))

## [0.8.0](https://github.com/ptah-sh/ptah-agent/compare/v0.7.0...v0.8.0) (2024-07-16)


### Features

* [#5](https://github.com/ptah-sh/ptah-agent/issues/5) download new release into a subfolder ([a972b80](https://github.com/ptah-sh/ptah-agent/commit/a972b80c3fa6654cb443c6a792e230e80d96ebdd))

## [0.7.0](https://github.com/ptah-sh/ptah-agent/compare/v0.6.0...v0.7.0) (2024-07-16)


### Features

* [#5](https://github.com/ptah-sh/ptah-agent/issues/5) try set executable flag ([bdcc887](https://github.com/ptah-sh/ptah-agent/commit/bdcc887c58e89b2dee5b1726a5add577f530e0ed))

## [0.6.0](https://github.com/ptah-sh/ptah-agent/compare/v0.5.1...v0.6.0) (2024-07-16)


### Features

* [#20](https://github.com/ptah-sh/ptah-agent/issues/20) add support for s3 credentials validation ([fd4ee54](https://github.com/ptah-sh/ptah-agent/commit/fd4ee541a73207e46cc75887be2e0633ce527105))
* [#22](https://github.com/ptah-sh/ptah-agent/issues/22) support backup tasks ([98d56b2](https://github.com/ptah-sh/ptah-agent/commit/98d56b2e20f6dbc14c79bd8024b154d770c2261b))

## [0.5.1](https://github.com/ptah-sh/ptah-agent/compare/v0.5.0...v0.5.1) (2024-07-04)


### Bug Fixes

* [#18](https://github.com/ptah-sh/ptah-agent/issues/18) match items by full name ([1aa7686](https://github.com/ptah-sh/ptah-agent/commit/1aa76865583d0eb5a167650fcad313ce8b35457f))

## [0.5.0](https://github.com/ptah-sh/ptah-agent/compare/v0.4.0...v0.5.0) (2024-07-04)


### Features

* [#17](https://github.com/ptah-sh/ptah-agent/issues/17) add support for the release command ([6f1f083](https://github.com/ptah-sh/ptah-agent/commit/6f1f08344b21680107546fda827c1422b512bf13))


### Bug Fixes

* [#15](https://github.com/ptah-sh/ptah-agent/issues/15) provide docker id to CreateRegistryAuthRes ([7a9b6fe](https://github.com/ptah-sh/ptah-agent/commit/7a9b6fe6c7c3dfd5ae5bfed775d5d18d895d3dad))

## [0.4.0](https://github.com/ptah-sh/ptah-agent/compare/v0.3.0...v0.4.0) (2024-07-02)


### Features

* [#13](https://github.com/ptah-sh/ptah-agent/issues/13) support deployments from the private registreis ([f0b6b19](https://github.com/ptah-sh/ptah-agent/commit/f0b6b1988579a2217c574b594df79317110556f6))

## [0.3.0](https://github.com/ptah-sh/ptah-agent/compare/v0.2.0...v0.3.0) (2024-07-01)


### Features

* [#5](https://github.com/ptah-sh/ptah-agent/issues/5) support self-upgrade command ([dda2a8b](https://github.com/ptah-sh/ptah-agent/commit/dda2a8b1995290c4e2a6536426c5ac92b4d63cdf))

## [0.2.0](https://github.com/ptah-sh/ptah-agent/compare/v0.1.0...v0.2.0) (2024-06-29)


### Features

* [#9](https://github.com/ptah-sh/ptah-agent/issues/9) improve naming of the resulting binary ([c434b5c](https://github.com/ptah-sh/ptah-agent/commit/c434b5c4d682728711c5a7a41f65d34295fa82d6))


### Bug Fixes

* [#9](https://github.com/ptah-sh/ptah-agent/issues/9) remove test step for now ([9b076ef](https://github.com/ptah-sh/ptah-agent/commit/9b076efa1e8c6eb6bf716468fea16ea634a5f1d5))
* [#9](https://github.com/ptah-sh/ptah-agent/issues/9) try fix ci ([da728ea](https://github.com/ptah-sh/ptah-agent/commit/da728eaf90c30e248eaa6b81fa406b0287f5dfbf))

## [0.1.0](https://github.com/ptah-sh/ptah-agent/compare/v0.0.1...v0.1.0) (2024-06-28)


### Features

* [#9](https://github.com/ptah-sh/ptah-agent/issues/9) add installation script ([deb738d](https://github.com/ptah-sh/ptah-agent/commit/deb738d63bdedd826cc2fb6d7360c144759c3da5))

## 0.0.1 (2024-06-27)


### Features

* [#1](https://github.com/ptah-sh/ptah-agent/issues/1) initialise the agent ([2355bc6](https://github.com/ptah-sh/ptah-agent/commit/2355bc638b467d9b6b1f3fdaceca8f583fb175f1))
* [#3](https://github.com/ptah-sh/ptah-agent/issues/3) support service creation ([b7e17e1](https://github.com/ptah-sh/ptah-agent/commit/b7e17e176c7834092d7e482403f7ae2cc93dcd00))
* [#4](https://github.com/ptah-sh/ptah-agent/issues/4) support caddy's config upload ([1f4f0fb](https://github.com/ptah-sh/ptah-agent/commit/1f4f0fbaa13d0ec837cf5fdaab3ff2919c048c9b))
* [#6](https://github.com/ptah-sh/ptah-agent/issues/6) support upading services ([07b04b1](https://github.com/ptah-sh/ptah-agent/commit/07b04b12599a89af2ae5d3933bf695e23a199259))
* [#7](https://github.com/ptah-sh/ptah-agent/issues/7) support service deletion ([6d37ab7](https://github.com/ptah-sh/ptah-agent/commit/6d37ab7c2807460a3b562f7d099dcbf35576c16c))


### Miscellaneous Chores

* [#2](https://github.com/ptah-sh/ptah-agent/issues/2) add ci/cd ([a4d9ca6](https://github.com/ptah-sh/ptah-agent/commit/a4d9ca65a269e07d54b665f63260ddfbc8e9f1d4))
