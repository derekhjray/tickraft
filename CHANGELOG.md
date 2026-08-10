# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - YYYY-MM-DD

### Added
- Initial open-source release of Tickraft.
- Scheduling engine with cron / interval / event / one-shot support.
- HTTP / TCP / ICMP / local script executors.
- Active probing (ICMP / TCP / HTTP) and Webhook passive listener.
- Alert rules engine with Webhook notification channel.
- Asset management with state machine and timeout detection.
- Built-in admin user with JWT authentication and API key management.

### Known Limitations
- Single-process All-in-One mode (no clustering).
- SQLite-only storage driver.
- Open-source quotas: 20 assets / 20 probes / 20 schedules / 5 remediations / 60s HTTP interval / 100k events/day.
