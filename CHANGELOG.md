
<a name="v1.1.2"></a>
# v1.1.2 (2015-07-27)

## :arrow_up: Dependency Updates

- **core**:
  - Update golifx to v0.0.5 ([eb255e6c](https://github.com/pdf/kodi-callback-daemon/commit/eb255e6cf372a4d946054a2ee1527da49d6d737d))


<a name="v1.1.1"></a>
# v1.1.1 (2015-07-26)

## :bug: Bug Fixes

- **lifx**:
  - Enable LIFX discovery polling and honor debug flag for logs ([0b10454b](https://github.com/pdf/kodi-callback-daemon/commit/0b10454b131d7378185716fb32d8f6b29a9e2ba9))


<a name="v1.1.0"></a>
# v1.1.0 (2015-07-26)

## :sparkles: Features

- **packaging**:
  - Add systemd unit ([4ef46666](https://github.com/pdf/kodi-callback-daemon/commit/4ef466663e10447a2ffcfd9897e725c93262b25a))
- **backends**:
  - Add LIFX backend ([a7514a06](https://github.com/pdf/kodi-callback-daemon/commit/a7514a06ab6bc0fb21e6a1c7d87c10331060f70c))


<a name="v1.0.4"></a>
# v1.0.4 (2015-07-05)

## :arrow_up: Dependency Updates

- **core**:
  - Update kodi_jsonrpc to v2.0.4 ([8476c080](https://github.com/pdf/kodi-callback-daemon/commit/8476c0802efad33e9f2be0e1580774b0eb33cb5d))


<a name="v1.0.3"></a>
# v1.0.3 (2015-07-05)

## :house: Housekeeping

- **build**:
  - Migrate from godeps to goop, update goxc config ([f8bc5a3a](https://github.com/pdf/kodi-callback-daemon/commit/f8bc5a3a8fbb99d81b8e777af22d353c51df36b6))

## :arrow_up: Dependency Updates

- **core**:
  - Update kodi lib to v2.0.3 ([0db4f84b](https://github.com/pdf/kodi-callback-daemon/commit/0db4f84b484b963f3b605591f6400965abdc7b2b))


<a name="v1.0.2"></a>
# v1.0.2 (2015-03-17)

## :house: Housekeeping

- **libs**:
  - Update kodi_jsonrpc to v2.0.0 ([fb1ce3ad](https://github.com/pdf/kodi-callback-daemon/commit/fb1ce3ad7adec8c186cb73898bbddbf22b4f9dd1))


<a name="v1.0.1"></a>
# v1.0.1 (2014-11-25)

## :house: Housekeeping

- **libs**:
  - Update kodi_jsonrpc to v1.0.3 for bugfixes ([f5103f81](https://github.com/pdf/kodi-callback-daemon/commit/f5103f8183256574ed15f21137e1e8c7dcf93d19))


<a name="v1.0.0"></a>
# v1.0.0 (2014-11-23)

## :house: Housekeeping

- **build**:
  - Add config for abe33/changelog-gen ([b5d4d358](https://github.com/pdf/kodi-callback-daemon/commit/b5d4d35810071b83467f08725078896a5205c6c2))  
    <br>TODO: Add contribution guidelines for commit messages
- **libs**:
  - Update to kodi_jsonrpc v1.0.1 ([cf6d5e78](https://github.com/pdf/kodi-callback-daemon/commit/cf6d5e78ada0d56b806ea67e91985e3bba90259d))
- **core**:
  - Rename from XBMC to Kodi, best to get this out of the way ([183f565a](https://github.com/pdf/kodi-callback-daemon/commit/183f565ade30119c11dfb6c58d68c799413d7bc4))

## Breaking Changes

- due to [183f565a](https://github.com/pdf/kodi-callback-daemon/commit/183f565ade30119c11dfb6c58d68c799413d7bc4), the change in package name requires users to uninstall the old package and replace it

