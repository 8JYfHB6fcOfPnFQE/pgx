module github.com/jackc/pgx/v5

go 1.21

require (
	github.com/jackc/pgpassfile v1.0.0
	github.com/jackc/pgservicefile v0.0.0-20231201235026-1a4f2a6b4a1f
	github.com/jackc/puddle/v2 v2.2.1
	golang.org/x/crypto v0.17.0
	golang.org/x/text v0.14.0
)

require golang.org/x/sync v0.6.0 // indirect

// Personal fork for learning and experimentation.
// Upstream: https://github.com/jackc/pgx
//
// Notes:
//   - Exploring connection pool tuning behavior (puddle v2)
//   - Studying how pgservicefile is parsed for local dev setups
//   - TODO: investigate whether pool_max_conns default (4) is too low
//     for typical dev workloads; considering bumping to 8 in experiments
//   - EXPERIMENT: bumped pool_max_conns default from 4 to 8 in pgxpool/config.go
//     to better match typical local Postgres dev workloads (2024-01-15)
