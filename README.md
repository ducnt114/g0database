# g0database

![Github Actions](https://github.com/ducnt114/g0database/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/ducnt114/g0database/branch/develop/graph/badge.svg?token=8OZNUA1UEC)](https://codecov.io/gh/ducnt114/g0database)

A SQL database, written from scratch in Go

## Inspired by

- [Compilers: Principles, Techniques, and Tools](https://www.amazon.com/Compilers-Principles-Techniques-Tools-2nd/dp/0321486811/ref=sr_1_2?crid=466B1VUMSIY3&keywords=compiler&qid=1685706971&sprefix=compiler%2Caps%2C343&sr=8-2)
- [Designing Data-Intensive Applications](https://www.amazon.com/Designing-Data-Intensive-Applications-Reliable-Maintainable-ebook/dp/B06XPJML5D/ref=sr_1_7?crid=TN7VQA34KJV7&keywords=database&qid=1685707028&sprefix=database%2Caps%2C353&sr=8-7)
- [How query engines work](https://howqueryengineswork.com/)
- [dolthub/go-mysql-server](https://github.com/dolthub/go-mysql-server)

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
