package g0database

import (
	"encoding/binary"
	"fmt"
	"io"
)

// MySQL command bytes
const (
	COM_QUIT    byte = 0x01
	COM_INIT_DB byte = 0x02
	COM_QUERY   byte = 0x03
	COM_PING    byte = 0x0e
)

// MySQL column types
const (
	MYSQL_TYPE_TINY     byte = 0x01 // BOOLEAN
	MYSQL_TYPE_LONG     byte = 0x03 // INT
	MYSQL_TYPE_LONGLONG byte = 0x08 // BIGINT
	MYSQL_TYPE_DATETIME byte = 0x0c
	MYSQL_TYPE_VARCHAR  byte = 0x0f
	MYSQL_TYPE_BLOB     byte = 0xfc // TEXT
)

// Capability flags
const (
	CLIENT_LONG_PASSWORD      uint32 = 0x00000001
	CLIENT_FOUND_ROWS         uint32 = 0x00000002
	CLIENT_LONG_FLAG          uint32 = 0x00000004
	CLIENT_CONNECT_WITH_DB    uint32 = 0x00000008
	CLIENT_PROTOCOL_41        uint32 = 0x00000200
	CLIENT_SECURE_CONNECTION  uint32 = 0x00008000
	CLIENT_PLUGIN_AUTH        uint32 = 0x00080000
	CLIENT_PLUGIN_AUTH_LENENC uint32 = 0x00200000
	CLIENT_DEPRECATE_EOF      uint32 = 0x01000000
)

// Server status flags
const (
	SERVER_STATUS_AUTOCOMMIT uint16 = 0x0002
)

// Character set
const (
	CHARSET_UTF8_GENERAL_CI byte = 0x21 // utf8_general_ci (33)
)

// readLengthEncodedInt reads a length-encoded integer from data
// Returns the value and the number of bytes consumed
func readLengthEncodedInt(data []byte) (uint64, int) {
	if len(data) == 0 {
		return 0, 0
	}

	switch data[0] {
	case 0xfb: // NULL
		return 0, 1
	case 0xfc: // 2-byte integer
		if len(data) < 3 {
			return 0, 0
		}
		return uint64(binary.LittleEndian.Uint16(data[1:3])), 3
	case 0xfd: // 3-byte integer
		if len(data) < 4 {
			return 0, 0
		}
		return uint64(data[1]) | uint64(data[2])<<8 | uint64(data[3])<<16, 4
	case 0xfe: // 8-byte integer
		if len(data) < 9 {
			return 0, 0
		}
		return binary.LittleEndian.Uint64(data[1:9]), 9
	default:
		if data[0] < 0xfb {
			return uint64(data[0]), 1
		}
		return 0, 0
	}
}

// writeLengthEncodedInt writes a length-encoded integer
func writeLengthEncodedInt(n uint64) []byte {
	switch {
	case n < 251:
		return []byte{byte(n)}
	case n < 1<<16:
		buf := make([]byte, 3)
		buf[0] = 0xfc
		binary.LittleEndian.PutUint16(buf[1:], uint16(n))
		return buf
	case n < 1<<24:
		buf := make([]byte, 4)
		buf[0] = 0xfd
		buf[1] = byte(n)
		buf[2] = byte(n >> 8)
		buf[3] = byte(n >> 16)
		return buf
	default:
		buf := make([]byte, 9)
		buf[0] = 0xfe
		binary.LittleEndian.PutUint64(buf[1:], n)
		return buf
	}
}

// readLengthEncodedString reads a length-encoded string from data
// Returns the string and number of bytes consumed
func readLengthEncodedString(data []byte) (string, int) {
	length, n := readLengthEncodedInt(data)
	if n == 0 {
		return "", 0
	}
	if len(data) < n+int(length) {
		return "", 0
	}
	return string(data[n : n+int(length)]), n + int(length)
}

// writeLengthEncodedString writes a length-encoded string
func writeLengthEncodedString(s string) []byte {
	lenBytes := writeLengthEncodedInt(uint64(len(s)))
	result := make([]byte, len(lenBytes)+len(s))
	copy(result, lenBytes)
	copy(result[len(lenBytes):], s)
	return result
}

// readNullTerminatedString reads a null-terminated string from data
// Returns the string and number of bytes consumed (including null terminator)
func readNullTerminatedString(data []byte) (string, int) {
	for i, b := range data {
		if b == 0 {
			return string(data[:i]), i + 1
		}
	}
	return string(data), len(data)
}

// writeNullTerminatedString writes a null-terminated string
func writeNullTerminatedString(s string) []byte {
	result := make([]byte, len(s)+1)
	copy(result, s)
	result[len(s)] = 0
	return result
}

// readPacket reads a MySQL packet from the reader
// Returns payload, sequence ID, and error
func readPacket(r io.Reader) ([]byte, uint8, error) {
	// Read 4-byte header
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, 0, err
	}

	// Extract length (3 bytes, little-endian)
	length := uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16
	seqID := header[3]

	// Read payload
	payload := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, 0, err
		}
	}

	return payload, seqID, nil
}

// writePacket writes a MySQL packet to the writer
func writePacket(w io.Writer, seqID uint8, payload []byte) error {
	length := len(payload)
	if length > 0xffffff {
		return fmt.Errorf("packet too large: %d bytes", length)
	}

	// Build header
	header := make([]byte, 4)
	header[0] = byte(length)
	header[1] = byte(length >> 8)
	header[2] = byte(length >> 16)
	header[3] = seqID

	// Write header and payload
	if _, err := w.Write(header); err != nil {
		return err
	}
	if length > 0 {
		if _, err := w.Write(payload); err != nil {
			return err
		}
	}
	return nil
}

// writeOKPacket creates an OK packet
func writeOKPacket(affectedRows, lastInsertID uint64) []byte {
	var buf []byte
	buf = append(buf, 0x00) // OK header
	buf = append(buf, writeLengthEncodedInt(affectedRows)...)
	buf = append(buf, writeLengthEncodedInt(lastInsertID)...)
	// Status flags (2 bytes) - autocommit
	buf = append(buf, byte(SERVER_STATUS_AUTOCOMMIT), byte(SERVER_STATUS_AUTOCOMMIT>>8))
	// Warnings (2 bytes)
	buf = append(buf, 0x00, 0x00)
	return buf
}

// writeERRPacket creates an ERR packet
func writeERRPacket(errorCode uint16, sqlState, message string) []byte {
	var buf []byte
	buf = append(buf, 0xff) // ERR header
	// Error code (2 bytes, little-endian)
	buf = append(buf, byte(errorCode), byte(errorCode>>8))
	// SQL state marker
	buf = append(buf, '#')
	// SQL state (5 bytes)
	if len(sqlState) < 5 {
		sqlState = sqlState + "     "[:5-len(sqlState)]
	}
	buf = append(buf, sqlState[:5]...)
	// Error message
	buf = append(buf, message...)
	return buf
}

// writeEOFPacket creates an EOF packet
func writeEOFPacket() []byte {
	var buf []byte
	buf = append(buf, 0xfe) // EOF header
	// Warnings (2 bytes)
	buf = append(buf, 0x00, 0x00)
	// Status flags (2 bytes)
	buf = append(buf, byte(SERVER_STATUS_AUTOCOMMIT), byte(SERVER_STATUS_AUTOCOMMIT>>8))
	return buf
}

// writeColumnDefinition creates a column definition packet
func writeColumnDefinition(schema, table, name string, colType byte, size uint32) []byte {
	var buf []byte

	// Catalog (always "def")
	buf = append(buf, writeLengthEncodedString("def")...)
	// Schema
	buf = append(buf, writeLengthEncodedString(schema)...)
	// Virtual table name
	buf = append(buf, writeLengthEncodedString(table)...)
	// Physical table name
	buf = append(buf, writeLengthEncodedString(table)...)
	// Virtual column name
	buf = append(buf, writeLengthEncodedString(name)...)
	// Physical column name
	buf = append(buf, writeLengthEncodedString(name)...)

	// Fixed length fields marker (0x0c)
	buf = append(buf, 0x0c)

	// Character set (2 bytes) - utf8_general_ci
	buf = append(buf, CHARSET_UTF8_GENERAL_CI, 0x00)

	// Column length (4 bytes)
	buf = append(buf,
		byte(size),
		byte(size>>8),
		byte(size>>16),
		byte(size>>24),
	)

	// Column type (1 byte)
	buf = append(buf, colType)

	// Flags (2 bytes) - no flags
	buf = append(buf, 0x00, 0x00)

	// Decimals (1 byte)
	buf = append(buf, 0x00)

	// Filler (2 bytes)
	buf = append(buf, 0x00, 0x00)

	return buf
}

// dataTypeToMySQLType converts g0database DataType to MySQL column type
func dataTypeToMySQLType(dt DataType) byte {
	switch dt {
	case DataTypeInt:
		return MYSQL_TYPE_LONG
	case DataTypeBigInt:
		return MYSQL_TYPE_LONGLONG
	case DataTypeVarchar:
		return MYSQL_TYPE_VARCHAR
	case DataTypeText:
		return MYSQL_TYPE_BLOB
	case DataTypeDateTime:
		return MYSQL_TYPE_DATETIME
	case DataTypeBoolean:
		return MYSQL_TYPE_TINY
	default:
		return MYSQL_TYPE_VARCHAR
	}
}

// writeRowData creates a text protocol row packet
func writeRowData(values []interface{}) []byte {
	var buf []byte
	for _, v := range values {
		if v == nil {
			buf = append(buf, 0xfb) // NULL marker
		} else {
			s := fmt.Sprintf("%v", v)
			buf = append(buf, writeLengthEncodedString(s)...)
		}
	}
	return buf
}
