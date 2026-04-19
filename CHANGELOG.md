# Changelog

All notable changes to Kymaros are documented in this file.

## [0.6.8] - 2026-04-19

### Added

**Plugin architecture**
- Public extension points for backup adapters, notifiers, and API routes (`pkg/plugin/`)
- Extracted `main()` into importable `pkg/app.Run()` for external Go modules
- `RegisterGlobalRoutes()` for plugin route registration via `init()`

**Dashboard — complete UI redesign**
- New design system with Inter + JetBrains Mono fonts, oklch color tokens, and dense SRE-optimized layout
- Sidebar navigation with logo, version display, active states, and external links
- TopBar with dynamic breadcrumbs, live status indicator, search trigger, notifications bell, and theme toggle
- Command palette (Cmd+K) with navigation, actions, dynamic test list, and external resources
- Dark/light theme toggle with localStorage persistence
- Toast notification system for transient action feedback
- Notifications drawer (Shift+N) with persistent feed, unread indicators, and mark-all-read
- Dashboard page: hero stats grid, minimalist score trend chart, test distribution stacked bar, filterable tests table (All/Pass/Partial/Fail), clickable alerts and upcoming tests, manual refresh button
- Report detail page: validation breakdown with 6 weighted progress bars, health checks table with full messages, restore completeness resource cards, pod logs accordion, Kubernetes events table, score history chart with run marker, test timeline
- Reports page: historical runs with search, date range filters (7d/30d/90d), status filters, client-side pagination
- Tests page: CRUD with structured form (create/edit), trigger button, delete with confirmation, View YAML dialog, kebab action menu
- Settings page: tabbed layout (Provider, Sandbox, Notifications, API, About), API token management with save/clear/kubectl instructions, "Managed via Helm values" alerts for read-only sections
- Keyboard shortcuts: Cmd+K (palette), Shift+N (notifications), ESC (close dialogs)

**Examples**
- 5 ready-to-use examples: wordpress-mysql, postgresql-ha, redis-sentinel, microservices-app, compliance-dora
- Each includes app manifests, Velero backup schedule, RestoreTest, and HealthCheckPolicy
- DORA compliance example with Grafana dashboard panel

### Fixed
- Upcoming tests displayed "just now" instead of future time ("in 3h", "in 1d")
- Score trend chart tooltip showed unrounded floats (75.333 → 75)
- RTO values > 24h (from corrupted historical data) now display as "—" instead of "12343h 18m"
- Average RTO calculation excludes invalid values > 24h
- `HandleTestNotification` returns 501 instead of misleading 200
- `fmt.Sprintf` without format verbs removed from `pkg/app/run.go`
- HTTP shutdown uses 10s timeout instead of unbounded `context.Background()`
- `flag.Parse()` guarded with `flag.Parsed()` for safe external module imports
- All displayed scores use `Math.round()` consistently
- Trigger error messages now show detailed backend response instead of generic "Failed to trigger"

### Changed
- Dashboard auto-refresh replaced with manual Refresh button (eliminates visual flicker)
- Schedule cron expressions displayed in human-readable format ("Daily at 3:30 UTC")
- `formatCron`, `formatRelativeTime`, `formatDuration`, `formatRTO` centralized in `lib/utils.ts`

## [0.6.7] - 2026-04-11

### Added
- GPG chart signing for supply chain security
- ArtifactHub verified publisher configuration
- Dashboard branding update

### Fixed
- Switch GPG key to RSA4096 for Helm compatibility

## [0.6.6] - 2026-04-08

### Added
- Gateway API (HTTPRoute) support for ingress

## [0.6.5] - 2026-04-08

### Security
- Comprehensive security hardening pass

### Fixed
- CORS tests for explicit origin handling

## [0.6.4] - 2026-04-08

### Added
- Global notification defaults via Helm values
- Helm values.schema.json for input validation

### Security
- Fix CVE-2026-33186 (grpc) and CVE-2026-24051 (otel)

### Changed
- Migrate install instructions from OCI to Helm repo

## [0.6.3] - 2026-04-08

### Added
- Bearer token auth for write API endpoints

## [0.6.2] - 2026-04-08

- Initial public release
