# g0database - Project Roadmap

## Overview

**Goal**: Build an in-memory MySQL-compatible database for unit testing pipelines.

**Constraints**:
- Go standard library only (no 3rd-party dependencies)
- Lightweight and fast startup
- MySQL wire protocol compatible (clients can connect via `mysql` CLI)
- Focus on unit-test use cases

---

## Current State (January 2025)

| Component | Status | Notes |
|-----------|--------|-------|
| **Lexer** | Done | Tokenizes SQL statements correctly |
| **Parser** | Partial | SELECT/CREATE work; INSERT/UPDATE/DELETE stubbed |
| **Executor** | Stubbed | Returns placeholder responses |
| **In-Memory Storage** | Not Started | No data persistence mechanism |
| **MySQL Protocol** | Not Started | No network layer |

### Existing Files
- `lexer.go` - SQL tokenization
- `parser.go` - AST construction
- `executor.go` - Command execution (stub)
- `command.go` - Command types and interfaces
- `model.go` - Schema/Table/Column models
- `token.go` - Token definitions
- `data_source.go` - CSV data source (prototype)
- `optimizer.go` - Empty placeholder

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│              MySQL Client (mysql-cli)           │
└─────────────────┬───────────────────────────────┘
                  │ TCP :3306
┌─────────────────▼───────────────────────────────┐
│           MySQL Protocol Handler                │
│   - Handshake & Authentication                  │
│   - Query Request/Response                      │
│   - Result Set Encoding                         │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│              SQL Engine                         │
│   Lexer → Parser → Planner → Executor           │
└─────────────────┬───────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│         In-Memory Storage Engine                │
│   - Table Storage (rows, columns)               │
│   - Index Support (B-tree or hash)              │
│   - Transaction Support (optional)              │
└─────────────────────────────────────────────────┘
```

---

## Development Phases

### Phase 1: In-Memory Storage Engine
Build the foundation for storing and retrieving data.

**Tasks**:
1. Design table storage structure (rows as slices/maps)
2. Implement basic operations: Insert, Select, Update, Delete
3. Add primary key support
4. Add simple indexing (hash-based for equality lookups)
5. Write unit tests for storage operations

**Deliverables**:
- `storage.go` - Table and row management
- `index.go` - Index structures
- `storage_test.go` - Comprehensive tests

---

### Phase 2: Complete SQL Parser
Finish parsing all required SQL statements.

**Tasks**:
1. Complete INSERT statement parsing
2. Complete UPDATE statement parsing
3. Complete DELETE statement parsing
4. Implement WHERE clause parsing (conditions, AND/OR)
5. Add support for common expressions (literals, comparisons)
6. Add ORDER BY, LIMIT parsing
7. Add basic JOIN parsing (INNER JOIN)

**Deliverables**:
- Enhanced `parser.go`
- `expression.go` - Expression evaluation
- `parser_test.go` - Full test coverage

---

### Phase 3: Query Executor
Connect parser output to storage engine.

**Tasks**:
1. Implement SELECT execution with WHERE filtering
2. Implement INSERT execution
3. Implement UPDATE execution
4. Implement DELETE execution
5. Implement CREATE TABLE / DROP TABLE
6. Add expression evaluation engine
7. Implement ORDER BY sorting
8. Implement LIMIT/OFFSET

**Deliverables**:
- Enhanced `executor.go`
- `evaluator.go` - Expression evaluation
- `executor_test.go` - Integration tests

---

### Phase 4: MySQL Protocol Implementation
Make the database accessible via MySQL clients.

**Tasks**:
1. Implement TCP server (net package)
2. Implement MySQL handshake protocol
3. Implement authentication (mysql_native_password)
4. Implement COM_QUERY command handling
5. Implement result set encoding (text protocol)
6. Implement error packet encoding
7. Handle connection lifecycle

**MySQL Protocol Reference**:
- Handshake: Server greeting → Client auth → OK/ERR
- Query: COM_QUERY packet → Result Set / OK / ERR
- Result Set: Column count → Column definitions → Rows → EOF

**Deliverables**:
- `server.go` - TCP server and connection handling
- `protocol.go` - MySQL packet encoding/decoding
- `auth.go` - Authentication handling

---

### Phase 5: Integration & Polish
Make it production-ready for unit tests.

**Tasks**:
1. Add graceful shutdown
2. Add connection pooling support
3. Add common MySQL functions (NOW(), UUID(), etc.)
4. Add SHOW commands (SHOW TABLES, SHOW DATABASES)
5. Add DESCRIBE/EXPLAIN support
6. Performance optimization
7. Documentation and examples

**Deliverables**:
- `functions.go` - Built-in SQL functions
- `README.md` - Usage documentation
- Example test cases

---

## SQL Feature Support (Target)

### Must Have (Phase 1-4)
- [ ] SELECT with WHERE, ORDER BY, LIMIT
- [ ] INSERT (single and multi-row)
- [ ] UPDATE with WHERE
- [ ] DELETE with WHERE
- [ ] CREATE TABLE / DROP TABLE
- [ ] Primary keys
- [ ] Basic data types: INT, BIGINT, VARCHAR, TEXT, DATETIME, BOOLEAN
- [ ] Comparison operators: =, <>, <, >, <=, >=
- [ ] Logical operators: AND, OR, NOT
- [ ] NULL handling: IS NULL, IS NOT NULL

### Nice to Have (Phase 5+)
- [ ] INNER JOIN
- [ ] Aggregations: COUNT, SUM, AVG, MIN, MAX
- [ ] GROUP BY / HAVING
- [ ] Subqueries
- [ ] AUTO_INCREMENT
- [ ] UNIQUE constraints
- [ ] Transactions (BEGIN, COMMIT, ROLLBACK)
- [ ] CREATE INDEX

### Out of Scope
- Foreign key constraints
- Triggers
- Stored procedures
- Views
- Full-text search
- Replication

---

## Technical Notes

### MySQL Protocol Basics

**Packet Structure**:
```
┌──────────┬──────────┬──────────────────┐
│ 3 bytes  │ 1 byte   │ N bytes          │
│ Length   │ Seq ID   │ Payload          │
└──────────┴──────────┴──────────────────┘
```

**Handshake Flow**:
1. Server → Client: Handshake packet (protocol version, server version, auth challenge)
2. Client → Server: Handshake response (username, auth response, database)
3. Server → Client: OK packet or ERR packet

**Query Flow**:
1. Client → Server: COM_QUERY packet (command byte + SQL string)
2. Server → Client: Result Set OR OK/ERR packet

### In-Memory Storage Design

```go
type Row struct {
    Values []interface{}  // Column values in order
}

type Table struct {
    Def   *TableDef
    Rows  []*Row
    pkCol int  // Primary key column index
}

type Database struct {
    Name   string
    Tables map[string]*Table
}
```

**Note**: No indexes - all lookups use simple table scans for simplicity.

---

## References

- [MySQL Protocol Documentation](https://dev.mysql.com/doc/dev/mysql-server/latest/PAGE_PROTOCOL.html)
- [MySQL Packet Format](https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_basic_packets.html)
- [MySQL Text Protocol](https://dev.mysql.com/doc/dev/mysql-server/latest/page_protocol_com_query.html)

---

## Progress Tracking

### Phase 1: Storage Engine ✅ COMPLETE
- [x] Table structure design
- [x] Row insert operation
- [x] Row select operation (table scan)
- [x] Row update operation
- [x] Row delete operation
- [x] Primary key uniqueness (via table scan)
- [x] Unit tests (28 tests passing)

### Phase 2: Parser Completion
- [ ] INSERT parsing
- [ ] UPDATE parsing
- [ ] DELETE parsing
- [ ] WHERE clause parsing
- [ ] Expression parsing
- [ ] ORDER BY / LIMIT
- [ ] Unit tests

### Phase 3: Executor
- [ ] SELECT execution
- [ ] INSERT execution
- [ ] UPDATE execution
- [ ] DELETE execution
- [ ] CREATE/DROP TABLE
- [ ] Expression evaluation
- [ ] Integration tests

### Phase 4: MySQL Protocol
- [ ] TCP server
- [ ] Handshake protocol
- [ ] Authentication
- [ ] COM_QUERY handling
- [ ] Result set encoding
- [ ] Error handling
- [ ] Connection tests

### Phase 5: Polish
- [ ] Graceful shutdown
- [ ] SHOW commands
- [ ] Built-in functions
- [ ] Documentation
- [ ] Example usage
