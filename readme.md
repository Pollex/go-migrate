## Go Migrate

A simple Postgres migration tool

Using this command the user can apply or undo database migrations.
The second parameter can be used either give a target migration or migrate relative to the current migration.

To migrate to the latest version, omit the second parameter:
'go-migrate -d postgres://root:root@localhost:5432/database ./migrations'

To migrate to the second migration, supply an integer as second parameter:
'go-migrate -d postgres://root:root@localhost:5432/database ./migrations 2'

To undo the last 3 migrations, supply a relative integer (prefixed with + or -):
'go-migrate -d postgres://root:root@localhost:5432/database ./migrations -3'

### Install

```
go install github.com/pollex/go-migrate
```
