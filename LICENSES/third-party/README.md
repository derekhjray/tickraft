# Third-Party Dependency Licenses

This directory retains the license texts of all **direct dependencies** declared in the tickraft open-source repository's `go.mod`, in order to comply with each license's requirement to retain the license text and copyright notice.

> **Maintenance rule**: When adding a new direct dependency, copy its LICENSE file into this directory named `<dependency>-LICENSE.txt`, and update the list below accordingly.

## Direct Dependency License List

| Dependency | Version | License | AGPLv3 Compatible | License File |
|------------|---------|---------|-------------------|--------------|
| github.com/BurntSushi/toml | v1.6.0 | MIT | Yes | [toml-LICENSE.txt](./toml-LICENSE.txt) |
| github.com/bytedance/sonic | v1.15.2 | Apache 2.0 | Yes | [sonic-LICENSE.txt](./sonic-LICENSE.txt) |
| github.com/cloudwego/hertz | v0.10.5 | Apache 2.0 | Yes | [hertz-LICENSE.txt](./hertz-LICENSE.txt) |
| github.com/expr-lang/expr | v1.17.8 | MIT | Yes | [expr-LICENSE.txt](./expr-LICENSE.txt) |
| github.com/fsnotify/fsnotify | v1.10.1 | BSD 3-Clause | Yes | [fsnotify-LICENSE.txt](./fsnotify-LICENSE.txt) |
| github.com/golang-jwt/jwt/v5 | v5.3.1 | MIT | Yes | [golang-jwt-jwt-v5-LICENSE.txt](./golang-jwt-jwt-v5-LICENSE.txt) |
| github.com/hertz-contrib/http2 | v0.1.8 | Apache 2.0 | Yes | [hertz-contrib-http2-LICENSE.txt](./hertz-contrib-http2-LICENSE.txt) |
| github.com/spf13/cobra | v1.10.2 | Apache 2.0 | Yes | [cobra-LICENSE.txt](./cobra-LICENSE.txt) |
| go.etcd.io/bbolt | v1.5.0 | MIT | Yes | [bbolt-LICENSE.txt](./bbolt-LICENSE.txt) |
| go.uber.org/zap | v1.28.0 | MIT | Yes | [zap-LICENSE.txt](./zap-LICENSE.txt) |
| golang.org/x/crypto | v0.54.0 | BSD 3-Clause | Yes | [golang-x-crypto-LICENSE.txt](./golang-x-crypto-LICENSE.txt) |
| golang.org/x/net | v0.57.0 | BSD 3-Clause | Yes | [golang-x-net-LICENSE.txt](./golang-x-net-LICENSE.txt) |
| gopkg.in/yaml.v3 | v3.0.1 | MIT / Apache 2.0 | Yes | [yaml-v3-LICENSE.txt](./yaml-v3-LICENSE.txt) |
| gorm.io/driver/sqlite | v1.6.0 | MIT | Yes | [gorm-driver-sqlite-LICENSE.txt](./gorm-driver-sqlite-LICENSE.txt) |
| gorm.io/gorm | v1.31.2 | MIT | Yes | [gorm-LICENSE.txt](./gorm-LICENSE.txt) |

## License Compatibility Notes

- **MIT / Apache 2.0 / BSD 2-Clause / BSD 3-Clause / ISC**: Permissive licenses, fully compatible with AGPLv3, permitted for use.
- **MPL 2.0**: Weak copyleft license (file-level copyleft), compatible with AGPLv3, permitted for use.
- **GPLv2 / SSPL / Proprietary**: Incompatible with AGPLv3, prohibited.

All direct dependencies listed above are under permissive or weak copyleft licenses, satisfying the open-source repository's license compliance requirements with no copyleft contamination risk.

## License Text Source

All license texts are copied verbatim from the corresponding dependency module's root directory (`LICENSE` / `LICENSE.txt` / `COPYING` / `License` file) in the Go Module Cache (`$GOMODCACHE`), without any modification, preserving the original copyright notice and full license text.

## Transitive Dependency Review

Indirect dependencies (`// indirect`) are reviewed quarterly based on `go.sum` and `go mod graph` output to ensure that the transitive dependency chain does not unexpectedly introduce incompatible licenses.

For automated license scanning, it is recommended to use [`go-licenses`](https://github.com/google/go-licenses) to generate a full dependency license report.
