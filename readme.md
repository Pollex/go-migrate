## Go Migrate

A simple Postgres migration tool

## How to create migrations

A single migration exists of two files: an up migration and a down migration. These files simply contain the SQL to be executed.

The file must have the following format:

```
<index>_<a-simple-description>.<direction>.sql

1_initial.up.sql
1_initial.down.sql

2_add_user_lastname.up.sql
2_add_user_lastname.down.sql

```

Indexes must be sequential and can not have gaps. (i.e. `1 2 4 5` is invalid because 3 is missing!)

## How to apply migrations

Using this command the user can apply or undo database migrations.
The second parameter can be used either give a target migration or migrate relative to the current migration.

To migrate to the latest version, omit the second parameter:

```
go-migrate -d postgres://root:root@localhost:5432/database ./migrations
```

To migrate to the second migration, supply an integer as second parameter:

```
go-migrate -d postgres://root:root@localhost:5432/database ./migrations 2
```

To undo the last 3 migrations, supply a relative integer (prefixed with + or -):

```
go-migrate -d postgres://root:root@localhost:5432/database ./migrations -3
```

## How to install

```
go install github.com/pollex/go-migrate
```
