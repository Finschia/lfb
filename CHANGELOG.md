<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes.
"State Machine Breaking" for breaking the AppState

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Features
* (app) Revise bech32 prefix cosmos to link and tlink

### Improvements
* (sdk) Use fastcache for inter block cache and iavl cache
* (sdk) Enable signature verification cache
* (ostracon) Apply asynchronous receiving reactor

### Bug Fixes

### Breaking Changes
* (sdk) (auth) [\#16](https://github.com/line/lfb/pull/16) Introduce sig block height for the new replay protection
* (ostracon/sdk) [\#26](https://github.com/line/lfb/pull/26) Use vrf-based consensus, address string treatment
* (global) [\#27](https://github.com/line/lfb/pull/27) Use lbm-sdk instead of lfb-sdk

## [gaia v4.0.4] - 2021-03-15
Initial lfb is based on the tendermint v0.34.9+, cosmos-sdk v0.42.0+, gaia v4.0.4

* (tendermint) [v0.34.9](https://github.com/tendermint/tendermint/releases/tag/v0.34.9).
* (cosmos-sdk) [v0.42.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.42.0).
* (gaia) [v4.0.4](https://github.com/cosmos/gaia/releases/tag/v4.0.4).

Please refer [CHANGELOG_OF_GAIA_v4.0.4](https://github.com/cosmos/gaia/blob/v4.0.4/CHANGELOG.md)
<!-- Release links -->
