# g0database

![Github Actions](https://github.com/ducnt114/g0database/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/gh/ducnt114/g0database/branch/develop/graph/badge.svg?token=8OZNUA1UEC)](https://codecov.io/gh/ducnt114/g0database)

A MySQL-like database, written from scratch in Go

## Inspired by

- [Compilers: Principles, Techniques, and Tools](https://www.amazon.com/Compilers-Principles-Techniques-Tools-2nd/dp/0321486811/ref=sr_1_2?crid=466B1VUMSIY3&keywords=compiler&qid=1685706971&sprefix=compiler%2Caps%2C343&sr=8-2)
- [Designing Data-Intensive Applications](https://www.amazon.com/Designing-Data-Intensive-Applications-Reliable-Maintainable-ebook/dp/B06XPJML5D/ref=sr_1_7?crid=TN7VQA34KJV7&keywords=database&qid=1685707028&sprefix=database%2Caps%2C353&sr=8-7)
- [How query engines work](https://howqueryengineswork.com/)
- [dolthub/go-mysql-server](https://github.com/dolthub/go-mysql-server)

## Documentation

See more in [docs](./docs/)

---

## Learning Path: Building a SQL Database

If you want to understand how SQL databases work internally, here are the key topics to study:

### 1. Compiler Theory & SQL Parsing

- **Lexical Analysis (Lexer/Tokenizer)** - Breaking SQL strings into tokens (keywords, identifiers, literals, operators)
- **Syntax Analysis (Parser)** - Building an Abstract Syntax Tree (AST) from tokens
- **Grammar & BNF Notation** - Understanding SQL grammar rules
- **Recursive Descent Parsing** - A common technique for parsing SQL
- **Pratt Parsing** - For handling operator precedence in expressions

**Resources:**
- [Crafting Interpreters](https://craftinginterpreters.com/) - Excellent intro to lexers and parsers
- [Writing a SQL Parser](https://tomassetti.me/parsing-sql/)

### 2. Query Processing

- **Query Planning** - Converting AST to execution plan
- **Query Optimization** - Cost-based optimization, join ordering, index selection
- **Predicate Pushdown** - Moving filters closer to data source
- **Expression Evaluation** - Evaluating WHERE clauses, functions, operators

**Resources:**
- [How Query Engines Work](https://howqueryengineswork.com/)
- [CMU Database Systems Course](https://15445.courses.cs.cmu.edu/)

### 3. Storage Engines

- **Page-Based Storage** - How databases organize data in fixed-size pages
- **B-Tree / B+Tree** - The most common index structure for databases
- **LSM Tree (Log-Structured Merge Tree)** - Used by LevelDB, RocksDB, Cassandra
- **SSTable (Sorted String Table)** - Immutable sorted files used with LSM
- **Write-Ahead Log (WAL)** - Durability through logging before writes
- **MVCC (Multi-Version Concurrency Control)** - How databases handle concurrent access
- **Buffer Pool Management** - Caching pages in memory

**Resources:**
- [Database Internals](https://www.databass.dev/) by Alex Petrov
- [Designing Data-Intensive Applications](https://dataintensive.net/) - Chapter 3

### 4. Transaction Management

- **ACID Properties** - Atomicity, Consistency, Isolation, Durability
- **Isolation Levels** - Read Uncommitted, Read Committed, Repeatable Read, Serializable
- **Locking** - Row locks, table locks, deadlock detection
- **2PC (Two-Phase Commit)** - Distributed transaction protocol

### 5. Network Protocol

- **MySQL Wire Protocol** - Packet format, handshake, query/response
- **Connection Management** - Connection pooling, session state
- **Result Set Encoding** - How query results are serialized

**Resources:**
- [MySQL Protocol Documentation](https://dev.mysql.com/doc/dev/mysql-server/latest/PAGE_PROTOCOL.html)

### 6. Data Structures & Algorithms

- **Hash Tables** - For hash indexes and hash joins
- **Binary Search** - For B-tree traversal
- **Sorting Algorithms** - For ORDER BY, merge joins
- **Bloom Filters** - For efficient negative lookups

### 7. Recommended Reading Order

1. **Start here:** [Crafting Interpreters](https://craftinginterpreters.com/) - Learn lexer/parser basics
2. **Then:** [How Query Engines Work](https://howqueryengineswork.com/) - Understand query execution
3. **Deep dive:** [Database Internals](https://www.databass.dev/) - Storage engine details
4. **Big picture:** [Designing Data-Intensive Applications](https://dataintensive.net/) - Distributed systems context

### 8. Open Source Databases to Study

| Database | Language | Notable For |
|----------|----------|-------------|
| [SQLite](https://sqlite.org/src/doc/trunk/README.md) | C | Simple, well-documented, single-file |
| [CockroachDB](https://github.com/cockroachdb/cockroach) | Go | Distributed SQL, great docs |
| [TiDB](https://github.com/pingcap/tidb) | Go | MySQL compatible, distributed |
| [DuckDB](https://github.com/duckdb/duckdb) | C++ | Analytical (OLAP), columnar |
| [go-mysql-server](https://github.com/dolthub/go-mysql-server) | Go | MySQL protocol implementation |

### 9. Key Concepts Summary

```
SQL Query Lifecycle:
┌──────────┐    ┌────────┐    ┌──────────┐    ┌───────────┐    ┌─────────┐
│  SQL     │───▶│ Lexer  │───▶│  Parser  │───▶│ Optimizer │───▶│Executor │
│  String  │    │(Tokens)│    │  (AST)   │    │  (Plan)   │    │(Results)│
└──────────┘    └────────┘    └──────────┘    └───────────┘    └─────────┘
                                                                    │
                                                                    ▼
                                                            ┌─────────────┐
                                                            │   Storage   │
                                                            │   Engine    │
                                                            │ (B-Tree/LSM)│
                                                            └─────────────┘
```
