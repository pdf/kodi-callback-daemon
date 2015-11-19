
<a name="v1.4.0"></a>
# v1.4.0 (2015-11-20)

## :sparkles: Features

- **lifx**:
  - Boblight proxy support ([a03e5b70](https://github.com/pdf/kodi-callback-daemon/commit/a03e5b702adca6820bb3110dc661a7a8ab52fd32))

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.3.6 ([9bcff4af](https://github.com/pdf/kodi-callback-daemon/commit/9bcff4afbca6a97e3604b35f5f108b25fd823c66))  
    <br>Use the newly exposed color averaging from golifx for more accurate hues


<a name="v1.3.5"></a>
# v1.3.5 (2015-11-10)

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.3.4 ([986c42fd](https://github.com/pdf/kodi-callback-daemon/commit/986c42fddca93cfb86d44ddda73d85f3ffb24346))


<a name="v1.3.4"></a>
# v1.3.4 (2015-11-02)

## :house: Housekeeping

- **build**:
  - Fix goxc exclude dirs for Go 1.5 vendoring ([3e7472fa](https://github.com/pdf/kodi-callback-daemon/commit/3e7472fa8cfbe5a30381f3b133138f69f2acbb44))

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.3.1 ([163b4b86](https://github.com/pdf/kodi-callback-daemon/commit/163b4b86b2a29c8c40c52bf1cbfe88309bebe350))


<a name="v1.3.3"></a>
# v1.3.3 (2015-10-18)

## :house: Housekeeping

- **build**:
  - Update goxc task list ([6a69a8cb](https://github.com/pdf/kodi-callback-daemon/commit/6a69a8cb35b10e4ce25db86dd5211e3cea63933e))
  - Migrate from goop to glide for dependency management ([d5135dad](https://github.com/pdf/kodi-callback-daemon/commit/d5135dad2a30265403242a7dd0c033f060f5093c))  
    <br>Enables Go 1.5 native vendoring

## :arrow_up: Dependency Updates

- **core**:
  - Bump kodi_jsonrpc to v2.0.5 ([cc912ed4](https://github.com/pdf/kodi-callback-daemon/commit/cc912ed46a0254f8fad9d89d3e79caef5b0584f7))
  - Bump golifx to v0.3.1 ([a95e1ccb](https://github.com/pdf/kodi-callback-daemon/commit/a95e1ccb302f42c3a5dc4f17b3e46106f68d107f))


<a name="v1.3.2"></a>
# v1.3.2 (2015-10-16)

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.3.0 ([364e0bdf](https://github.com/pdf/kodi-callback-daemon/commit/364e0bdfede8e2c3971022cd935940c830c8443f))


<a name="v1.3.1"></a>
# v1.3.1 (2015-09-23)

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.2.2 ([3863567f](https://github.com/pdf/kodi-callback-daemon/commit/3863567fcb0d6a8161b5818d868195b6780ca397))


<a name="v1.3.0"></a>
# v1.3.0 (2015-09-12)

## :sparkles: Features

- **LIFX**:
  - Add group support ([70aafb2f](https://github.com/pdf/kodi-callback-daemon/commit/70aafb2f04286d3140371230ba21a0b1dfa3a615))

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.2.0 ([6617c7b0](https://github.com/pdf/kodi-callback-daemon/commit/6617c7b07cb3fd2919b8d6fc2929c60b3edfa33b))


<a name="v1.2.0"></a>
# v1.2.0 (2015-08-27)

## :sparkles: Features

- **core**:
  - Add media types to `Player.OnPause` and `Player.OnStop` ([0c290a54](https://github.com/pdf/kodi-callback-daemon/commit/0c290a5455261933b21e12c01146cba9d66b8df9), [#2](https://github.com/pdf/kodi-callback-daemon/issues/2))


<a name="v1.1.5"></a>
# v1.1.5 (2015-08-08)

## :bug: Bug Fixes

- **packaging**:
  - Add systemd unit to package via goxc ([6e6d8ba0](https://github.com/pdf/kodi-callback-daemon/commit/6e6d8ba0421ed12220dc4103838c77d8c6513e68))


<a name="v1.1.4"></a>
# v1.1.4 (2015-08-01)

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.1.2 ([ac009963](https://github.com/pdf/kodi-callback-daemon/commit/ac009963bcee5010fa70bed097c1cefc5907106b))


<a name="v1.1.3"></a>
# v1.1.3 (2015-07-30)

## :arrow_up: Dependency Updates

- **core**:
  - Bump golifx to v0.1.1 and enable reliable mode ([9edfc521](https://github.com/pdf/kodi-callback-daemon/commit/9edfc5217c2aab9823b5211547f88600c099d51d))


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

