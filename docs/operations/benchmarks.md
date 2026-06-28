# Benchmarking

Gogo includes benchmarks for framework paths that are performance-sensitive in
real applications:

- Router matching.
- Route reversing.
- ORM select compilation.
- ORM insert compilation.
- ORM row scanning.
- Migration autodetection on a small app.
- Admin changelist planning.
- Serializer validation.
- Queue publishing.
- Queue worker task execution with the in-memory broker and backend.

Run the benchmark suite with:

```bash
make bench
```

Run a focused benchmark with:

```bash
go test -run '^$' -bench BenchmarkRouterMatch -benchmem ./benchmarks
```

Do not commit local benchmark numbers as pass/fail thresholds. Workstation CPU,
thermal state, Go patch version, background load, and virtualization can all
change results. Use benchmarks for relative comparisons on the same machine or
in the same CI runner class.

For release comparisons:

1. Run the benchmark on the base commit.
2. Run the benchmark on the candidate commit.
3. Use the same Go version and machine class.
4. Compare allocations as well as time.
5. Investigate large regressions before release.

Benchmark fixtures must stay small, deterministic, and free of external
services. Integration performance testing for PostgreSQL, Redis, and RabbitMQ
belongs in deployment-specific load tests.
