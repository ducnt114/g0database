package g0database

import (
	"crypto/rand"
	"encoding/binary"
)

const (
	// Default server version string
	DefaultServerVersion = "8.0.0-g0database"

	// Auth plugin name
	AuthPluginMysqlNativePassword = "mysql_native_password"
)

// generateScramble generates a 20-byte random scramble for authentication
func generateScramble() []byte {
	scramble := make([]byte, 20)
	rand.Read(scramble)
	// Avoid null bytes which can cause issues
	for i := range scramble {
		if scramble[i] == 0 {
			scramble[i] = 1
		}
	}
	return scramble
}

// buildHandshakeV10 builds the server greeting packet (HandshakeV10)
func buildHandshakeV10(connID uint32, scramble []byte, serverVersion string) []byte {
	// Calculate capability flags
	capabilities := CLIENT_LONG_PASSWORD |
		CLIENT_FOUND_ROWS |
		CLIENT_LONG_FLAG |
		CLIENT_CONNECT_WITH_DB |
		CLIENT_PROTOCOL_41 |
		CLIENT_SECURE_CONNECTION |
		CLIENT_PLUGIN_AUTH

	var buf []byte

	// Protocol version (1 byte): always 10
	buf = append(buf, 0x0a)

	// Server version (null-terminated)
	buf = append(buf, writeNullTerminatedString(serverVersion)...)

	// Connection ID (4 bytes, little-endian)
	connIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(connIDBytes, connID)
	buf = append(buf, connIDBytes...)

	// Auth plugin data part 1 (8 bytes) - first part of scramble
	buf = append(buf, scramble[:8]...)

	// Filler (1 byte)
	buf = append(buf, 0x00)

	// Capability flags lower 2 bytes
	buf = append(buf, byte(capabilities), byte(capabilities>>8))

	// Character set (1 byte) - utf8_general_ci
	buf = append(buf, CHARSET_UTF8_GENERAL_CI)

	// Status flags (2 bytes) - autocommit
	buf = append(buf, byte(SERVER_STATUS_AUTOCOMMIT), byte(SERVER_STATUS_AUTOCOMMIT>>8))

	// Capability flags upper 2 bytes
	buf = append(buf, byte(capabilities>>16), byte(capabilities>>24))

	// Auth plugin data length (1 byte) - 21 for 20-byte scramble + null
	buf = append(buf, 21)

	// Reserved (10 bytes of zeros)
	buf = append(buf, make([]byte, 10)...)

	// Auth plugin data part 2 (12 bytes + null terminator = 13 bytes)
	buf = append(buf, scramble[8:]...)
	buf = append(buf, 0x00) // null terminator for scramble

	// Auth plugin name (null-terminated)
	buf = append(buf, writeNullTerminatedString(AuthPluginMysqlNativePassword)...)

	return buf
}

// HandshakeResponse holds parsed client handshake response
type HandshakeResponse struct {
	Capabilities  uint32
	MaxPacketSize uint32
	CharacterSet  uint8
	Username      string
	AuthResponse  []byte
	Database      string
	AuthPlugin    string
}

// parseHandshakeResponse parses client's handshake response
func parseHandshakeResponse(data []byte) (*HandshakeResponse, error) {
	if len(data) < 32 {
		return nil, ErrInvalidPacket
	}

	resp := &HandshakeResponse{}
	pos := 0

	// Capability flags (4 bytes)
	resp.Capabilities = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	// Max packet size (4 bytes)
	resp.MaxPacketSize = binary.LittleEndian.Uint32(data[pos:])
	pos += 4

	// Character set (1 byte)
	resp.CharacterSet = data[pos]
	pos++

	// Reserved (23 bytes of zeros)
	pos += 23

	// Username (null-terminated)
	username, n := readNullTerminatedString(data[pos:])
	resp.Username = username
	pos += n

	// Auth response
	if resp.Capabilities&CLIENT_PLUGIN_AUTH_LENENC != 0 {
		// Length-encoded auth response
		length, lenBytes := readLengthEncodedInt(data[pos:])
		pos += lenBytes
		if pos+int(length) <= len(data) {
			resp.AuthResponse = data[pos : pos+int(length)]
			pos += int(length)
		}
	} else if resp.Capabilities&CLIENT_SECURE_CONNECTION != 0 {
		// 1-byte length + auth response
		if pos < len(data) {
			length := int(data[pos])
			pos++
			if pos+length <= len(data) {
				resp.AuthResponse = data[pos : pos+length]
				pos += length
			}
		}
	} else {
		// Null-terminated auth response
		authStr, n := readNullTerminatedString(data[pos:])
		resp.AuthResponse = []byte(authStr)
		pos += n
	}

	// Database (null-terminated, if CLIENT_CONNECT_WITH_DB is set)
	if resp.Capabilities&CLIENT_CONNECT_WITH_DB != 0 && pos < len(data) {
		database, n := readNullTerminatedString(data[pos:])
		resp.Database = database
		pos += n
	}

	// Auth plugin name (null-terminated, if CLIENT_PLUGIN_AUTH is set)
	if resp.Capabilities&CLIENT_PLUGIN_AUTH != 0 && pos < len(data) {
		plugin, _ := readNullTerminatedString(data[pos:])
		resp.AuthPlugin = plugin
	}

	return resp, nil
}

// ErrInvalidPacket indicates an invalid packet format
var ErrInvalidPacket = ErrInvalidType // Reuse existing error
