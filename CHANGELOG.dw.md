# Changelog

All notable changes to this project will be documented in this file.
These images are published to the `docker.ihs.demonware.net/platform-dbtech/vitess-lite:` docker repository.

## Unreleased

## [v20.0.8-percona80-bc0d257409] - 2026-06-02
### Fixed
- [DBT-2110](https://dev.activision.com/jira/browse/DBT-2110)
  - Backport connpool fixes from [arthurschreiber/vitess#6](https://github.com/arthurschreiber/vitess/pull/6)
  - Fix race condition when waiting for connection ([#18713](https://github.com/vitessio/vitess/pull/18713))
  - Bump hang detection timeout to 1 second ([#18722](https://github.com/vitessio/vitess/pull/18722))
  - Fix connection leak during idle connection reopen ([#18967](https://github.com/vitessio/vitess/pull/18967))

## [v20.0.8-percona80-0ad4853597] - 2026-01-09
### Fixed
- [DBT-1810](https://dev.activision.com/jira/browse/DBT-1810): Backport "Remove shard targeted keyspace from database qualifier on the query" from upstream [fix](https://github.com/vitessio/vitess/pull/17503)
