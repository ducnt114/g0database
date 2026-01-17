# g0database - Project Roadmap

## Overview

**Goal**: Build an in-memory MySQL-compatible database for unit testing pipelines.

**Constraints**:
- Go standard library only (no 3rd-party dependencies)
- Lightweight and fast startup
- MySQL wire protocol compatible (clients can connect via `mysql` CLI)
- Focus on unit-test use cases

---

## Current State (January 17, 2026)

| Component | Status | Notes |
|-----------|--------|-------|
| **Lexer** | Done | Tokenizes SQL statements (note: quoted strings have issues) |
| **Parser** | Done | Full CRUD parsing with WHERE, ORDER BY, LIMIT |
| **Executor** | Done | Executes all CRUD operations with structured results |
| **In-Memory Storage** | Done | Table storage with CRUD operations |
| **MySQL Protocol** | Done | TCP server, handshake, query execution |

### Test Summary
- **Total Tests**: 100 passing
- **Storage Tests**: 28 tests
- **Parser Tests**: 18 tests
- **Executor Tests**: 37 tests
- **Protocol Tests**: 9 tests
- **Server Tests**: 8 tests

### Existing Files

**Core Components:**
- `lexer.go` - SQL tokenization (keywords, identifiers, operators)
- `parser.go` - AST construction (SELECT, INSERT, UPDATE, DELETE, CREATE, DROP)
- `executor.go` - Command execution with Engine integration
- `evaluator.go` - WHERE clause evaluation and predicate building
- `storage.go` - In-memory table and row management (Engine, Database, Table, Row)

**MySQL Protocol:**
- `server.go` - TCP server and connection handling
- `protocol.go` - MySQL packet encoding/decoding
- `auth.go` - Handshake and authentication

**Support Files:**
- `command.go` - Command types, interfaces, and CommandResult with structured data
- `model.go` - Data types (DataType, Column)
- `token.go` - Token definitions
- `errors.go` - Error definitions (ErrTableNotFound, ErrColumnNotFound, etc.)
- `data_source.go` - CSV data source (prototype)
- `optimizer.go` - Empty placeholder

**Test Files:**
- `lexer_test.go` - Lexer unit tests
- `parser_test.go` - Parser unit tests (18 tests)
- `storage_test.go` - Storage engine unit tests (28 tests)
- `executor_test.go` - Executor integration tests (37 tests)
- `protocol_test.go` - Protocol unit tests (9 tests)
- `server_test.go` - Server integration tests (8 tests)
- `data_source_test.go` - CSV data source tests

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
- [x] SELECT with WHERE, ORDER BY, LIMIT
- [x] INSERT (single row)
- [x] UPDATE with WHERE
- [x] DELETE with WHERE
- [x] CREATE TABLE / DROP TABLE
- [x] Primary keys
- [x] Basic data types: INT, BIGINT, VARCHAR, TEXT, DATETIME, BOOLEAN
- [x] Comparison operators: =, <>, <, >, <=, >=
- [x] Logical operators: AND, OR
- [ ] NULL handling: IS NULL, IS NOT NULL
- [ ] INSERT (multi-row)
- [ ] NOT operator

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

### Phase 2: Parser Completion ✅ COMPLETE
- [x] INSERT parsing
- [x] UPDATE parsing
- [x] DELETE parsing
- [x] WHERE clause parsing (with AND/OR)
- [x] Value parsing (NULL, bool, int, float, string)
- [x] ORDER BY / LIMIT
- [x] DROP TABLE parsing
- [x] Unit tests (18 parser tests)

### Phase 3: Executor ✅ COMPLETE
- [x] SELECT execution (with WHERE, ORDER BY, LIMIT)
- [x] INSERT execution (with/without column names)
- [x] UPDATE execution (with WHERE filtering)
- [x] DELETE execution (with WHERE filtering)
- [x] CREATE/DROP TABLE execution
- [x] Expression evaluation (evaluator.go)
- [x] Predicate building from WHERE clauses
- [x] All comparison operators (=, <>, <, >, <=, >=)
- [x] AND/OR logic support
- [x] ORDER BY sorting (ASC/DESC, multiple columns)
- [x] LIMIT support
- [x] Result formatting
- [x] Integration tests (37 executor tests, 83 total)

### Phase 4: MySQL Protocol ✅ COMPLETE
- [x] TCP server (server.go)
- [x] Handshake protocol (auth.go)
- [x] Authentication (mysql_native_password, accepts any password)
- [x] COM_QUERY handling
- [x] COM_INIT_DB (USE database)
- [x] COM_PING
- [x] COM_QUIT
- [x] Result set encoding (protocol.go)
- [x] Error packet encoding with MySQL error codes
- [x] OK packet encoding
- [x] EOF packet encoding
- [x] Length-encoded integers/strings
- [x] Column definition packets
- [x] SHOW DATABASES / SHOW TABLES (basic)
- [x] SET commands (ignored)
- [x] SELECT @@version
- [x] Enhanced CommandResult with structured data
- [x] Protocol tests (9 tests)
- [x] Server integration tests (8 tests)
- [x] Documentation (PHASE4_MYSQL_PROTOCOL.md)

### Phase 5: Polish
- [ ] Graceful shutdown
- [x] SHOW commands (basic support added in Phase 4)
- [ ] Built-in functions (NOW(), UUID(), etc.)
- [ ] DESCRIBE/EXPLAIN support
- [x] Documentation (docs/ folder)
- [ ] Example usage / README update
