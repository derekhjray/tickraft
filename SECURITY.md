# Security Policy

> Version: V1.0 | Updated: 2026-08-04 | Status: active

This document describes how to report security vulnerabilities for the Tickraft open-source repository and what support reporters and users can expect.

## Supported Versions

Tickraft is currently in active development. Security fixes are provided for the supported release line below. Pre-release, unreleased, and end-of-life versions are not eligible for separate security fixes, although reports are still welcome and will be assessed on a best-effort basis.

| Version   | Security Fixes                                         |
| --------- | ------------------------------------------------------ |
| 1.0.x     | Supported (security fixes are backported to this line) |
| < 1.0     | Not eligible for separate security fixes               |
| Unreleased (development branches / pre-release tags) | Not eligible for separate security fixes (report still welcome) |

## Reporting a Vulnerability

If you believe you have discovered a security vulnerability in Tickraft, please report it responsibly.

- **Report privately** to **security@tickraft.com**. Do **not** open a public GitHub Issue for security vulnerabilities.
- Please include the following information so we can reproduce and assess the issue efficiently:
  - A clear description of the vulnerability and the affected component.
  - Step-by-step reproduction instructions (minimal proof of concept if possible).
  - The affected version, commit hash, or branch.
  - Potential impact and attack scenarios.
  - A suggested fix or mitigation, if you have one.

### Response Timeline

- **Acknowledgement**: within **72 hours** of receiving the report.
- **Initial assessment**: within **7 days**, including a preliminary severity rating and planned next steps.
- **Coordinated disclosure**: after a fix has been developed, validated, and released, we will publish a security advisory and credit the reporter (unless they prefer to remain anonymous).

### Disclosure Rules

- Do **not** disclose the vulnerability publicly or to third parties until a fix is available and coordinated disclosure has been agreed upon with the maintainers.
- Do **not** access or modify data that does not belong to you, and do not degrade the availability of services for other users while researching the issue.

## Scope

### In Scope

- Vulnerabilities in the **Tickraft open-source repository code**, including the Go backend (`cmd/`, `internal/`, `pkg/`) and the open-source frontend kernel (`web/packages/core/`, `web/packages/features/`) shipped in this repository.
- Security-relevant defects in build scripts, configuration templates, and bundled deployment manifests that are maintained in this repository.
- Issues that compromise confidentiality, integrity, or availability of a Tickraft deployment when used as documented.

### Out of Scope

- Vulnerabilities in **third-party dependencies** — please report them upstream to the maintainer of the affected package. Once a fix is available upstream, we will update the dependency.
- Issues arising from **self-inflicted misconfiguration**, unsafe deployment practices, or deviation from documented deployment guidance.
- **Social engineering**, phishing, or physical attacks against Tickraft users, maintainers, or infrastructure.
- Vulnerabilities requiring an already-compromised environment or attacker-controlled privileged credentials to exploit.
- Theoretical issues without a concrete, reproducible attack path.

## Safe Harbor

Tickraft welcomes good-faith security research that helps improve the security of the project and its users. We consider security research conducted in accordance with this policy to be authorized and welcomed.

- We will **not** take legal action against reporters who act in good faith, respect this policy, and avoid harming users or disrupting services.
- In return, we ask reporters to honor the coordinated disclosure process, avoid accessing or exposing user data, and refrain from degrading service availability during research.

If you are unsure whether your planned research falls within this policy, please contact **security@tickraft.com** before proceeding.

## Contact

- Security reports: **security@tickraft.com**
- For all other (non-security) questions, please use the public channels described in `CONTRIBUTING.md`.
