package g0database

import (
	"bytes"
	"testing"
)

func TestReadWriteLengthEncodedInt(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
	}{
		{"zero", 0},
		{"small", 100},
		{"max_1byte", 250},
		{"min_2byte", 251},
		{"medium", 1000},
		{"max_2byte", 65535},
		{"min_3byte", 65536},
		{"large_3byte", 1000000},
		{"max_3byte", 16777215},
		{"min_8byte", 16777216},
		{"large", 1234567890},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := writeLengthEncodedInt(tt.value)
			decoded, n := readLengthEncodedInt(encoded)

			if n != len(encoded) {
				t.Errorf("bytes consumed = %d, want %d", n, len(encoded))
			}
			if decoded != tt.value {
				t.Errorf("decoded = %d, want %d", decoded, tt.value)
			}
		})
	}
}

func TestReadLengthEncodedIntEdgeCases(t *testing.T) {
	// Empty data
	val, n := readLengthEncodedInt([]byte{})
	if n != 0 || val != 0 {
		t.Errorf("empty: got val=%d, n=%d, want 0, 0", val, n)
	}

	// NULL marker
	val, n = readLengthEncodedInt([]byte{0xfb})
	if n != 1 || val != 0 {
		t.Errorf("NULL: got val=%d, n=%d, want 0, 1", val, n)
	}
}

func TestReadWriteLengthEncodedString(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"empty", ""},
		{"short", "hello"},
		{"medium", "hello world, this is a longer string for testing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := writeLengthEncodedString(tt.s)
			decoded, n := readLengthEncodedString(encoded)

			if n != len(encoded) {
				t.Errorf("bytes consumed = %d, want %d", n, len(encoded))
			}
			if decoded != tt.s {
				t.Errorf("decoded = %q, want %q", decoded, tt.s)
			}
		})
	}
}

func TestReadWriteNullTerminatedString(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"empty", ""},
		{"short", "hello"},
		{"with_spaces", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := writeNullTerminatedString(tt.s)
			decoded, n := readNullTerminatedString(encoded)

			if n != len(tt.s)+1 {
				t.Errorf("bytes consumed = %d, want %d", n, len(tt.s)+1)
			}
			if decoded != tt.s {
				t.Errorf("decoded = %q, want %q", decoded, tt.s)
			}
			if encoded[len(encoded)-1] != 0 {
				t.Errorf("last byte should be null terminator")
			}
		})
	}
}

func TestReadWritePacket(t *testing.T) {
	tests := []struct {
		name    string
		seqID   uint8
		payload []byte
	}{
		{"empty", 0, []byte{}},
		{"small", 1, []byte("hello")},
		{"medium", 5, bytes.Repeat([]byte("x"), 1000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := writePacket(&buf, tt.seqID, tt.payload)
			if err != nil {
				t.Fatalf("writePacket error: %v", err)
			}

			payload, seqID, err := readPacket(&buf)
			if err != nil {
				t.Fatalf("readPacket error: %v", err)
			}

			if seqID != tt.seqID {
				t.Errorf("seqID = %d, want %d", seqID, tt.seqID)
			}
			if !bytes.Equal(payload, tt.payload) {
				t.Errorf("payload mismatch")
			}
		})
	}
}

func TestWriteOKPacket(t *testing.T) {
	packet := writeOKPacket(5, 10)

	if packet[0] != 0x00 {
		t.Errorf("OK packet should start with 0x00, got 0x%02x", packet[0])
	}

	// Verify affected rows (5) is encoded after header
	affectedRows, n := readLengthEncodedInt(packet[1:])
	if affectedRows != 5 {
		t.Errorf("affectedRows = %d, want 5", affectedRows)
	}

	// Verify last insert ID (10)
	lastInsertID, _ := readLengthEncodedInt(packet[1+n:])
	if lastInsertID != 10 {
		t.Errorf("lastInsertID = %d, want 10", lastInsertID)
	}
}

func TestWriteERRPacket(t *testing.T) {
	packet := writeERRPacket(1064, "42000", "Syntax error")

	if packet[0] != 0xff {
		t.Errorf("ERR packet should start with 0xff, got 0x%02x", packet[0])
	}

	// Error code (little-endian)
	errorCode := uint16(packet[1]) | uint16(packet[2])<<8
	if errorCode != 1064 {
		t.Errorf("errorCode = %d, want 1064", errorCode)
	}

	// SQL state marker
	if packet[3] != '#' {
		t.Errorf("SQL state marker should be '#', got %c", packet[3])
	}

	// SQL state
	sqlState := string(packet[4:9])
	if sqlState != "42000" {
		t.Errorf("sqlState = %q, want %q", sqlState, "42000")
	}

	// Error message
	message := string(packet[9:])
	if message != "Syntax error" {
		t.Errorf("message = %q, want %q", message, "Syntax error")
	}
}

func TestWriteEOFPacket(t *testing.T) {
	packet := writeEOFPacket()

	if packet[0] != 0xfe {
		t.Errorf("EOF packet should start with 0xfe, got 0x%02x", packet[0])
	}

	// Should be 5 bytes total
	if len(packet) != 5 {
		t.Errorf("EOF packet length = %d, want 5", len(packet))
	}
}

func TestWriteColumnDefinition(t *testing.T) {
	packet := writeColumnDefinition("testdb", "users", "name", MYSQL_TYPE_VARCHAR, 255)

	// Should start with catalog "def"
	catalog, n := readLengthEncodedString(packet)
	if catalog != "def" {
		t.Errorf("catalog = %q, want %q", catalog, "def")
	}

	// Schema
	schema, m := readLengthEncodedString(packet[n:])
	if schema != "testdb" {
		t.Errorf("schema = %q, want %q", schema, "testdb")
	}
	n += m

	// Virtual table
	table, m := readLengthEncodedString(packet[n:])
	if table != "users" {
		t.Errorf("table = %q, want %q", table, "users")
	}
	n += m

	// Physical table (same)
	_, m = readLengthEncodedString(packet[n:])
	n += m

	// Virtual column name
	colName, m := readLengthEncodedString(packet[n:])
	if colName != "name" {
		t.Errorf("column name = %q, want %q", colName, "name")
	}
	n += m

	// Physical column name (same)
	_, m = readLengthEncodedString(packet[n:])
	n += m

	// Fixed length fields marker
	if packet[n] != 0x0c {
		t.Errorf("fixed fields marker = 0x%02x, want 0x0c", packet[n])
	}
}

func TestDataTypeToMySQLType(t *testing.T) {
	tests := []struct {
		dataType DataType
		expected byte
	}{
		{DataTypeInt, MYSQL_TYPE_LONG},
		{DataTypeBigInt, MYSQL_TYPE_LONGLONG},
		{DataTypeVarchar, MYSQL_TYPE_VARCHAR},
		{DataTypeText, MYSQL_TYPE_BLOB},
		{DataTypeDateTime, MYSQL_TYPE_DATETIME},
		{DataTypeBoolean, MYSQL_TYPE_TINY},
	}

	for _, tt := range tests {
		t.Run(string(tt.dataType), func(t *testing.T) {
			result := dataTypeToMySQLType(tt.dataType)
			if result != tt.expected {
				t.Errorf("dataTypeToMySQLType(%s) = 0x%02x, want 0x%02x", tt.dataType, result, tt.expected)
			}
		})
	}
}

func TestWriteRowData(t *testing.T) {
	// Test with mixed values including NULL
	row := writeRowData([]interface{}{1, "hello", nil, true})

	// First value: "1"
	s, n := readLengthEncodedString(row)
	if s != "1" {
		t.Errorf("first value = %q, want %q", s, "1")
	}

	// Second value: "hello"
	s, m := readLengthEncodedString(row[n:])
	if s != "hello" {
		t.Errorf("second value = %q, want %q", s, "hello")
	}
	n += m

	// Third value: NULL (0xfb)
	if row[n] != 0xfb {
		t.Errorf("NULL marker = 0x%02x, want 0xfb", row[n])
	}
	n++

	// Fourth value: "true"
	s, _ = readLengthEncodedString(row[n:])
	if s != "true" {
		t.Errorf("fourth value = %q, want %q", s, "true")
	}
}
