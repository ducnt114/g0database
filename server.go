package g0database

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Addr          string // Address to listen on (default ":3306")
	ServerVersion string // Server version string (default "8.0.0-g0database")
}

// Server represents the MySQL protocol server
type Server struct {
	config   ServerConfig
	listener net.Listener
	engine   *Engine
	parser   Parser

	mu     sync.Mutex
	connID uint32
	conns  map[uint32]*Connection
	closed bool
}

// NewServer creates a new MySQL protocol server
func NewServer(config ServerConfig, engine *Engine) *Server {
	if config.Addr == "" {
		config.Addr = ":3306"
	}
	if config.ServerVersion == "" {
		config.ServerVersion = DefaultServerVersion
	}

	return &Server{
		config: config,
		engine: engine,
		parser: NewParser(),
		conns:  make(map[uint32]*Connection),
	}
}

// Start starts the server and listens for connections
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = listener

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed {
				return nil
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

// Close stops the server
func (s *Server) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()

	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Addr returns the server's listen address
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.config.Addr
}

// handleConnection handles a new client connection
func (s *Server) handleConnection(conn net.Conn) {
	// Generate connection ID
	connID := atomic.AddUint32(&s.connID, 1)

	// Create connection object
	c := &Connection{
		id:       connID,
		conn:     conn,
		server:   s,
		executor: NewExecutor(s.engine),
		seqID:    0,
	}

	// Register connection
	s.mu.Lock()
	s.conns[connID] = c
	s.mu.Unlock()

	// Handle connection
	defer func() {
		conn.Close()
		s.mu.Lock()
		delete(s.conns, connID)
		s.mu.Unlock()
	}()

	c.handle()
}

// Connection represents a client connection
type Connection struct {
	id       uint32
	conn     net.Conn
	server   *Server
	executor Executor
	seqID    uint8
	database string
}

// handle handles the connection lifecycle
func (c *Connection) handle() {
	// Perform handshake
	if err := c.doHandshake(); err != nil {
		return
	}

	// Process commands
	for {
		if err := c.readCommand(); err != nil {
			return
		}
	}
}

// doHandshake performs the MySQL handshake
func (c *Connection) doHandshake() error {
	// Generate scramble
	scramble := generateScramble()

	// Send server greeting
	greeting := buildHandshakeV10(c.id, scramble, c.server.config.ServerVersion)
	if err := writePacket(c.conn, c.seqID, greeting); err != nil {
		return err
	}
	c.seqID++

	// Read client response
	payload, seqID, err := readPacket(c.conn)
	if err != nil {
		return err
	}
	c.seqID = seqID + 1

	// Parse handshake response
	resp, err := parseHandshakeResponse(payload)
	if err != nil {
		c.writeError(1045, "28000", "Access denied")
		return err
	}

	// Set database if specified
	if resp.Database != "" {
		c.database = resp.Database
		if err := c.server.engine.UseDatabase(resp.Database); err != nil {
			// Create database if it doesn't exist
			c.server.engine.CreateDatabase(resp.Database)
			c.server.engine.UseDatabase(resp.Database)
		}
	}

	// Send OK packet (accept any authentication for testing purposes)
	if err := c.sendOK(0, 0); err != nil {
		return err
	}

	return nil
}

// readCommand reads and processes a single command
func (c *Connection) readCommand() error {
	payload, seqID, err := readPacket(c.conn)
	if err != nil {
		return err
	}
	c.seqID = seqID + 1

	if len(payload) == 0 {
		return io.EOF
	}

	cmdByte := payload[0]
	data := payload[1:]

	switch cmdByte {
	case COM_QUIT:
		return io.EOF

	case COM_INIT_DB:
		c.handleInitDB(string(data))

	case COM_QUERY:
		c.handleQuery(string(data))

	case COM_PING:
		c.sendOK(0, 0)

	default:
		c.writeError(1047, "08S01", fmt.Sprintf("Unknown command: %d", cmdByte))
	}

	return nil
}

// handleInitDB handles USE database command
func (c *Connection) handleInitDB(dbName string) {
	dbName = strings.TrimSpace(dbName)

	// Try to use existing database
	err := c.server.engine.UseDatabase(dbName)
	if err != nil {
		// Create database if it doesn't exist
		c.server.engine.CreateDatabase(dbName)
		err = c.server.engine.UseDatabase(dbName)
	}

	if err != nil {
		c.writeError(1049, "42000", fmt.Sprintf("Unknown database '%s'", dbName))
		return
	}

	c.database = dbName
	c.sendOK(0, 0)
}

// handleQuery handles COM_QUERY command
func (c *Connection) handleQuery(sql string) {
	sql = strings.TrimSpace(sql)

	// Handle special commands that aren't parsed by our parser
	upperSQL := strings.ToUpper(sql)

	// Handle SET commands (ignore them for now)
	if strings.HasPrefix(upperSQL, "SET ") {
		c.sendOK(0, 0)
		return
	}

	// Handle SELECT @@version, @@version_comment, etc.
	if strings.HasPrefix(upperSQL, "SELECT @@") || strings.Contains(upperSQL, "@@VERSION") {
		c.sendVersionResult()
		return
	}

	// Handle SHOW commands
	if strings.HasPrefix(upperSQL, "SHOW ") {
		c.handleShowCommand(upperSQL)
		return
	}

	// Parse SQL
	lexer := NewNaiveLexer()
	tokens, err := lexer.Analyze(sql)
	if err != nil {
		c.writeError(1064, "42000", "Syntax error: "+err.Error())
		return
	}

	cmd, err := c.server.parser.Parse(tokens)
	if err != nil {
		c.writeError(1064, "42000", "Parse error: "+err.Error())
		return
	}

	// Execute command
	result := c.executor.Execute(cmd)

	// Send response based on result type
	if result.Type == ResultTypeError {
		c.writeErrorFromResult(result)
		return
	}

	if result.Type == ResultTypeSelect {
		c.writeResultSet(result)
	} else {
		c.sendOK(uint64(result.AffectedRows), 0)
	}
}

// handleShowCommand handles SHOW commands
func (c *Connection) handleShowCommand(sql string) {
	if strings.Contains(sql, "DATABASES") {
		c.sendDatabasesResult()
	} else if strings.Contains(sql, "TABLES") {
		c.sendTablesResult()
	} else if strings.Contains(sql, "WARNINGS") {
		c.sendEmptyResult("Level", "Code", "Message")
	} else {
		c.sendOK(0, 0)
	}
}

// sendVersionResult sends a result set with version information
func (c *Connection) sendVersionResult() {
	columns := []ResultColumn{
		{Name: "@@version", Type: DataTypeVarchar, Size: 255},
	}
	rows := []ResultRow{
		{Values: []interface{}{c.server.config.ServerVersion}},
	}
	c.writeResultSetData(columns, rows)
}

// sendDatabasesResult sends a result set with database names
func (c *Connection) sendDatabasesResult() {
	columns := []ResultColumn{
		{Name: "Database", Type: DataTypeVarchar, Size: 255},
	}
	var rows []ResultRow
	for name := range c.server.engine.Databases {
		rows = append(rows, ResultRow{Values: []interface{}{name}})
	}
	c.writeResultSetData(columns, rows)
}

// sendTablesResult sends a result set with table names
func (c *Connection) sendTablesResult() {
	columns := []ResultColumn{
		{Name: "Tables_in_" + c.database, Type: DataTypeVarchar, Size: 255},
	}
	var rows []ResultRow
	if c.server.engine.Current != nil {
		for name := range c.server.engine.Current.Tables {
			rows = append(rows, ResultRow{Values: []interface{}{name}})
		}
	}
	c.writeResultSetData(columns, rows)
}

// sendEmptyResult sends an empty result set with given column names
func (c *Connection) sendEmptyResult(colNames ...string) {
	columns := make([]ResultColumn, len(colNames))
	for i, name := range colNames {
		columns[i] = ResultColumn{Name: name, Type: DataTypeVarchar, Size: 255}
	}
	c.writeResultSetData(columns, nil)
}

// writeResultSet writes a result set from CommandResult
func (c *Connection) writeResultSet(result CommandResult) {
	c.writeResultSetData(result.Columns, result.Rows)
}

// writeResultSetData writes a result set with given columns and rows
func (c *Connection) writeResultSetData(columns []ResultColumn, rows []ResultRow) {
	// Column count
	colCount := writeLengthEncodedInt(uint64(len(columns)))
	writePacket(c.conn, c.seqID, colCount)
	c.seqID++

	// Column definitions
	for _, col := range columns {
		colDef := writeColumnDefinition(
			c.database,
			"",
			col.Name,
			dataTypeToMySQLType(col.Type),
			uint32(col.Size),
		)
		writePacket(c.conn, c.seqID, colDef)
		c.seqID++
	}

	// EOF packet after columns
	writePacket(c.conn, c.seqID, writeEOFPacket())
	c.seqID++

	// Row data
	for _, row := range rows {
		rowData := writeRowData(row.Values)
		writePacket(c.conn, c.seqID, rowData)
		c.seqID++
	}

	// EOF packet after rows
	writePacket(c.conn, c.seqID, writeEOFPacket())
	c.seqID++
}

// sendOK sends an OK packet
func (c *Connection) sendOK(affectedRows, lastInsertID uint64) error {
	okPacket := writeOKPacket(affectedRows, lastInsertID)
	err := writePacket(c.conn, c.seqID, okPacket)
	c.seqID++
	return err
}

// writeError writes an error packet
func (c *Connection) writeError(code uint16, state, message string) {
	errPacket := writeERRPacket(code, state, message)
	writePacket(c.conn, c.seqID, errPacket)
	c.seqID++
}

// writeErrorFromResult writes an error packet from CommandResult
func (c *Connection) writeErrorFromResult(result CommandResult) {
	// Map common errors to MySQL error codes
	msg := result.Output
	code := uint16(1064) // Default syntax error
	state := "42000"

	if strings.Contains(msg, "table not found") {
		code = 1146
		state = "42S02"
	} else if strings.Contains(msg, "table already exists") {
		code = 1050
		state = "42S01"
	} else if strings.Contains(msg, "no database selected") {
		code = 1046
		state = "3D000"
	} else if strings.Contains(msg, "column not found") {
		code = 1054
		state = "42S22"
	} else if strings.Contains(msg, "database not found") {
		code = 1049
		state = "42000"
	}

	c.writeError(code, state, msg)
}
