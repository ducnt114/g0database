package g0database

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestServer_StartStop(t *testing.T) {
	engine := NewEngine()
	server := NewServer(ServerConfig{Addr: ":0"}, engine)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Close server
	err := server.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Wait for Start to return
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestServer_Handshake(t *testing.T) {
	engine := NewEngine()
	server := NewServer(ServerConfig{Addr: ":0"}, engine)

	go server.Start()
	defer server.Close()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Connect to server
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read server greeting
	greeting, seqID, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read greeting: %v", err)
	}

	if seqID != 0 {
		t.Errorf("Expected sequence ID 0, got %d", seqID)
	}

	// Verify greeting starts with protocol version 10
	if greeting[0] != 0x0a {
		t.Errorf("Expected protocol version 10, got %d", greeting[0])
	}

	// Verify server version string is present
	serverVersion, _ := readNullTerminatedString(greeting[1:])
	if serverVersion != DefaultServerVersion {
		t.Errorf("Expected server version %s, got %s", DefaultServerVersion, serverVersion)
	}
}

func TestServer_HandshakeWithAuth(t *testing.T) {
	engine := NewEngine()
	server := NewServer(ServerConfig{Addr: ":0"}, engine)

	go server.Start()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read server greeting
	_, _, err = readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read greeting: %v", err)
	}

	// Send client handshake response
	clientResponse := buildClientHandshakeResponse("root", "testdb")
	err = writePacket(conn, 1, clientResponse)
	if err != nil {
		t.Fatalf("Failed to send handshake response: %v", err)
	}

	// Read server response (should be OK)
	response, _, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read auth response: %v", err)
	}

	if response[0] != 0x00 {
		t.Errorf("Expected OK packet (0x00), got 0x%02x", response[0])
	}
}

func TestServer_Query_CreateTable(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("test")
	engine.UseDatabase("test")
	server := NewServer(ServerConfig{Addr: ":0"}, engine)

	go server.Start()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	conn := connectAndAuth(t, server.Addr())
	defer conn.Close()

	// Send CREATE TABLE query
	result := sendQuery(t, conn, "CREATE TABLE users (id INT, name VARCHAR(255))")

	// Should be OK packet
	if result[0] != 0x00 {
		t.Errorf("Expected OK packet, got 0x%02x", result[0])
	}
}

func TestServer_Query_Insert(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("test")
	engine.UseDatabase("test")

	// Create table with integer columns only (lexer has issues with quoted strings)
	tableDef := &TableDef{
		Name: "numbers",
		Columns: []ColumnDef{
			{Name: "id", Type: DataTypeInt},
			{Name: "value", Type: DataTypeInt},
		},
	}
	engine.Current.CreateTable(tableDef)

	server := NewServer(ServerConfig{Addr: ":0"}, engine)
	go server.Start()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	// Connect with "test" database
	conn, err := net.Dial("tcp", server.Addr())
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read greeting
	_, _, err = readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read greeting: %v", err)
	}

	// Send auth response with "test" database
	clientResponse := buildClientHandshakeResponse("root", "test")
	err = writePacket(conn, 1, clientResponse)
	if err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	// Read auth result
	_, _, err = readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read auth result: %v", err)
	}

	// Send INSERT query with integer values only
	result := sendQuery(t, conn, "INSERT INTO numbers VALUES (1, 100)")

	// Should be OK packet
	if result[0] != 0x00 {
		t.Errorf("Expected OK packet, got 0x%02x. Response: %v", result[0], result)
	}

	// Verify affected rows (only if OK packet)
	if result[0] == 0x00 {
		affectedRows, _ := readLengthEncodedInt(result[1:])
		if affectedRows != 1 {
			t.Errorf("Expected 1 affected row, got %d", affectedRows)
		}
	}
}

func TestServer_Query_Select(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("test")
	engine.UseDatabase("test")

	// Create table and insert data
	tableDef := &TableDef{
		Name: "users",
		Columns: []ColumnDef{
			{Name: "id", Type: DataTypeInt},
			{Name: "name", Type: DataTypeVarchar, Size: 255},
		},
	}
	engine.Current.CreateTable(tableDef)
	table := engine.Current.GetTable("users")
	table.Insert([]interface{}{1, "Alice"})
	table.Insert([]interface{}{2, "Bob"})

	server := NewServer(ServerConfig{Addr: ":0"}, engine)
	go server.Start()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	conn := connectAndAuth(t, server.Addr())
	defer conn.Close()

	// Send SELECT query
	seqID := sendQueryPacket(t, conn, "SELECT * FROM users")

	// Read column count
	colCountPacket, seqID, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read column count: %v", err)
	}
	colCount, _ := readLengthEncodedInt(colCountPacket)
	if colCount != 2 {
		t.Errorf("Expected 2 columns, got %d", colCount)
	}

	// Read column definitions
	for i := 0; i < int(colCount); i++ {
		_, seqID, err = readPacket(conn)
		if err != nil {
			t.Fatalf("Failed to read column definition %d: %v", i, err)
		}
	}

	// Read EOF after columns
	eofPacket, seqID, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read EOF: %v", err)
	}
	if eofPacket[0] != 0xfe {
		t.Errorf("Expected EOF packet, got 0x%02x", eofPacket[0])
	}

	// Read rows
	rowCount := 0
	for {
		packet, _, err := readPacket(conn)
		if err != nil {
			t.Fatalf("Failed to read row/EOF: %v", err)
		}
		if packet[0] == 0xfe {
			break // EOF
		}
		rowCount++
	}

	if rowCount != 2 {
		t.Errorf("Expected 2 rows, got %d", rowCount)
	}
	_ = seqID // suppress unused warning
}

func TestServer_Query_Error(t *testing.T) {
	engine := NewEngine()
	engine.CreateDatabase("test")
	engine.UseDatabase("test")
	server := NewServer(ServerConfig{Addr: ":0"}, engine)

	go server.Start()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	conn := connectAndAuth(t, server.Addr())
	defer conn.Close()

	// Query non-existent table
	result := sendQuery(t, conn, "SELECT * FROM nonexistent")

	// Should be ERR packet
	if result[0] != 0xff {
		t.Errorf("Expected ERR packet, got 0x%02x", result[0])
	}
}

func TestServer_Ping(t *testing.T) {
	engine := NewEngine()
	server := NewServer(ServerConfig{Addr: ":0"}, engine)

	go server.Start()
	defer server.Close()

	time.Sleep(100 * time.Millisecond)

	conn := connectAndAuth(t, server.Addr())
	defer conn.Close()

	// Send PING
	err := writePacket(conn, 0, []byte{COM_PING})
	if err != nil {
		t.Fatalf("Failed to send PING: %v", err)
	}

	// Read response
	response, _, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read PING response: %v", err)
	}

	if response[0] != 0x00 {
		t.Errorf("Expected OK packet, got 0x%02x", response[0])
	}
}

// Helper functions

func buildClientHandshakeResponse(username, database string) []byte {
	capabilities := CLIENT_LONG_PASSWORD |
		CLIENT_FOUND_ROWS |
		CLIENT_LONG_FLAG |
		CLIENT_CONNECT_WITH_DB |
		CLIENT_PROTOCOL_41 |
		CLIENT_SECURE_CONNECTION |
		CLIENT_PLUGIN_AUTH

	var buf bytes.Buffer

	// Capability flags (4 bytes)
	buf.WriteByte(byte(capabilities))
	buf.WriteByte(byte(capabilities >> 8))
	buf.WriteByte(byte(capabilities >> 16))
	buf.WriteByte(byte(capabilities >> 24))

	// Max packet size (4 bytes)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x01})

	// Character set (1 byte)
	buf.WriteByte(CHARSET_UTF8_GENERAL_CI)

	// Reserved (23 bytes)
	buf.Write(make([]byte, 23))

	// Username (null-terminated)
	buf.WriteString(username)
	buf.WriteByte(0)

	// Auth response length + response (empty for testing)
	buf.WriteByte(0)

	// Database (null-terminated)
	buf.WriteString(database)
	buf.WriteByte(0)

	// Auth plugin name (null-terminated)
	buf.WriteString(AuthPluginMysqlNativePassword)
	buf.WriteByte(0)

	return buf.Bytes()
}

func connectAndAuth(t *testing.T, addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Read greeting
	_, _, err = readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read greeting: %v", err)
	}

	// Send auth response
	clientResponse := buildClientHandshakeResponse("root", "test")
	err = writePacket(conn, 1, clientResponse)
	if err != nil {
		t.Fatalf("Failed to send auth: %v", err)
	}

	// Read auth result
	response, _, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read auth result: %v", err)
	}

	if response[0] != 0x00 {
		t.Fatalf("Auth failed: 0x%02x", response[0])
	}

	return conn
}

func sendQuery(t *testing.T, conn net.Conn, sql string) []byte {
	sendQueryPacket(t, conn, sql)

	response, _, err := readPacket(conn)
	if err != nil {
		t.Fatalf("Failed to read query response: %v", err)
	}

	return response
}

func sendQueryPacket(t *testing.T, conn net.Conn, sql string) uint8 {
	queryPacket := append([]byte{COM_QUERY}, []byte(sql)...)
	err := writePacket(conn, 0, queryPacket)
	if err != nil {
		t.Fatalf("Failed to send query: %v", err)
	}
	return 1
}
