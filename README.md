# g0database

![Github Actions](https://github.com/ducnt114/g0database/actions/workflows/go.yml/badge.svg)

A SQL database is written from scratch in Go

## Lexer

- Split raw sql command to token

raw sql: `select id, name, age from user where status = 'ACTIVE' and id > 10`

=> tokens: `["select", "id", ",", "name", ",", "age", "from", "user", "where", "status", "=", "'", "ACTIVE", "'", "and", "id", ">", "10"]`

## Parser

- Make Abstract syntax tree (AST) from list tokens

```text
root
 |- select
 |---- fields (id, name, age)
 |- from
 |---- tables (user)
 |- where
 |---- conditions
           |---- status
           |---- =
           |---- ACTIVE
       |- AND
           |---- id
           |---- >
           |---- 10
```
