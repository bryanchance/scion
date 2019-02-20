# Postgres

We use postgres from a docker container.
`./scion.sh topology` will automatically start it.
If you want to manually start it use `./tools/dc postgres up`.
To stop it use `./tools/dc postgres down`.

## Go Implementation

Currently we work against [jackc/pgx](https://github.com/jackc/pgx).
