package jsonparser

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

type JsonParser struct {
	data     string
	Integers []string
	Floating []string
}

const maxStartEndStringLen = 80

func NewParser(data string) *JsonParser {
	return &JsonParser{data: data}
}

func (json *JsonParser) IsValid() error {
	json.data = json.removeWS(json.data)

	tail, err := json.parseJson(json.data)
	if err != nil {
		return fmt.Errorf("cannot parse JSON: %s; unparsed tail: %q", err, json.makeErrorTrace(tail))
	}
	tail = json.removeWS(tail)
	if len(tail) > 0 {
		return fmt.Errorf("unexpected tail: %q", json.makeErrorTrace(tail))
	}
	return nil
}

func (json *JsonParser) removeWS(data string) string {
	if len(data) == 0 || data[0] > 0x20 {
		return data
	}
	if len(data) == 0 || data[0] != 0x20 && data[0] != 0x0A && data[0] != 0x09 && data[0] != 0x0D {
		return data
	}
	for i := 1; i < len(data); i++ {
		if data[i] != 0x20 && data[i] != 0x0A && data[i] != 0x09 && data[i] != 0x0D {
			return data[i:]
		}
	}
	return ""
}

func (json *JsonParser) makeErrorTrace(data string) string {
	if len(data) <= maxStartEndStringLen {
		return data
	}
	start := data[:40]
	end := data[len(data)-40:]
	return start + "..." + end
}

func (json *JsonParser) parseJson(data string) (string, error) {
	if len(data) == 0 {
		return data, fmt.Errorf("cannot parse empty string")
	}

	if data[0] == '{' {
		tail, err := json.parseObject(data[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse object: %s", err)
		}
		return tail, nil
	}
	if data[0] == '[' {
		tail, err := json.parseArray(data[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse array: %s", err)
		}
		return tail, nil
	}
	if data[0] == '"' {
		tail, err := json.parseString(data[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse string: %s", err)
		}
		return tail, nil
	}
	if data[0] == 't' {
		if len(data) < len("true") || data[:len("true")] != "true" {
			return data, fmt.Errorf("unexpected value found: %q", data)
		}
		return data[len("true"):], nil
	}
	if data[0] == 'f' {
		if len(data) < len("false") || data[:len("false")] != "false" {
			return data, fmt.Errorf("unexpected value found: %q", data)
		}
		return data[len("false"):], nil
	}
	if data[0] == 'n' {
		if len(data) < len("null") || data[:len("null")] != "null" {
			return data, fmt.Errorf("unexpected value found: %q", data)
		}
		return data[len("null"):], nil
	}

	numberBytes, tail, err := json.parseNumber(data)
	if err != nil {
		return tail, fmt.Errorf("cannot parse number: %s", err)
	}

	number := json.byteArrayToString(numberBytes)
	n := strings.IndexByte(number, '.')
	if n > 0 {
		json.Floating = append(json.Floating, number)
	} else {
		json.Integers = append(json.Integers, number)
	}

	return tail, nil
}

func (json *JsonParser) parseObject(data string) (string, error) {
	data = json.removeWS(data)
	if len(data) == 0 {
		return data, fmt.Errorf("missing '}'")
	}
	if data[0] == '}' {
		return data[1:], nil
	}

	for {
		var err error

		// Parse key.
		data = json.removeWS(data)
		if len(data) == 0 || data[0] != '"' {
			return data, fmt.Errorf(`cannot find opening '"" for object key`)
		}

		data, err = json.validateKey(data[1:])
		if err != nil {
			return data, fmt.Errorf("cannot parse object key: %s", err)
		}
		data = json.removeWS(data)
		if len(data) == 0 || data[0] != ':' {
			return data, fmt.Errorf("missing ':' after object key")
		}
		data = data[1:]

		// Parse value
		data = json.removeWS(data)
		data, err = json.parseJson(data)
		if err != nil {
			return data, fmt.Errorf("cannot parse object value: %s", err)
		}
		data = json.removeWS(data)
		if len(data) == 0 {
			return data, fmt.Errorf("unexpected end of object")
		}
		if data[0] == ',' {
			data = data[1:]
			continue
		}
		if data[0] == '}' {
			return data[1:], nil
		}
		return data, fmt.Errorf("missing ',' after object value")
	}
}

func (json *JsonParser) parseArray(data string) (string, error) {
	data = json.removeWS(data)
	if len(data) == 0 {
		return data, fmt.Errorf("missing ']'")
	}
	if data[0] == ']' {
		return data[1:], nil
	}

	for {
		var err error

		data = json.removeWS(data)
		data, err = json.parseJson(data)
		if err != nil {
			return data, fmt.Errorf("cannot parse array value: %s", err)
		}

		data = json.removeWS(data)
		if len(data) == 0 {
			return data, fmt.Errorf("unexpected end of array")
		}
		if data[0] == ',' {
			data = data[1:]
			continue
		}
		if data[0] == ']' {
			data = data[1:]
			return data, nil
		}
		return data, fmt.Errorf("missing ',' after array value")
	}
}

func (json *JsonParser) parseString(data string) (string, error) {
	// Try fast path - a string without escape sequences.
	if n := strings.IndexByte(data, '"'); n >= 0 && strings.IndexByte(data[:n], '\\') < 0 {
		return data[n+1:], nil
	}

	// Slow path - escape sequences are present.
	rs, tail, err := json.parseRawString(data)
	if err != nil {
		return tail, err
	}
	for {
		n := strings.IndexByte(rs, '\\')
		if n < 0 {
			return tail, nil
		}
		n++
		if n >= len(rs) {
			return tail, fmt.Errorf("BUG: parseRawString returned invalid string with trailing backslash: %q", rs)
		}
		ch := rs[n]
		rs = rs[n+1:]

		if ch < 0x20 {
			return tail, fmt.Errorf("string cannot contain control char 0x%02X", ch)
		}

		switch ch {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			// Valid escape sequences - see http://json.org/
			break
		case 'u':
			if len(rs) < 4 {
				return tail, fmt.Errorf(`too short escape sequence: \u%s`, rs)
			}
			xs := rs[:4]
			_, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				return tail, fmt.Errorf(`invalid escape sequence \u%s: %s`, xs, err)
			}
			rs = rs[4:]
		default:
			return tail, fmt.Errorf(`unknown escape sequence \%c`, ch)
		}
	}
}

func (json *JsonParser) parseNumber(data string) ([]byte, string, error) {
	var number []byte
	if len(data) == 0 {
		return number, data, fmt.Errorf("zero-length number")
	}
	if data[0] == '-' {
		number = append(number, '-')
		data = data[1:]
		if len(data) == 0 {
			return number, data, fmt.Errorf("missing number after minus")
		}
	}
	i := 0
	for i < len(data) {
		if data[i] < '0' || data[i] > '9' {
			break
		}
		number = append(number, data[i])
		i++
	}
	if i <= 0 {
		return number, data, fmt.Errorf("expecting 0..9 digit, got %c", data[0])
	}
	if data[0] == '0' && i != 1 {
		return number, data, fmt.Errorf("unexpected number starting from 0")
	}
	if i >= len(data) {
		return number, "", nil
	}
	if data[i] == '.' {
		// Validate fractional part
		number = append(number, data[i])
		data = data[i+1:]
		if len(data) == 0 {
			return number, data, fmt.Errorf("missing fractional part")
		}
		i = 0
		for i < len(data) {
			if data[i] < '0' || data[i] > '9' {
				break
			}
			number = append(number, data[i])
			i++
		}
		if i == 0 {
			return number, data, fmt.Errorf("expecting 0..9 digit in fractional part, got %c", data[0])
		}
		if i >= len(data) {
			return number, "", nil
		}
	}
	return number, data[i:], nil
}

func (json *JsonParser) parseRawString(data string) (string, string, error) {
	n := strings.IndexByte(data, '"')
	if n < 0 {
		return data, "", fmt.Errorf(`missing closing '"'`)
	}
	if n == 0 || data[n-1] != '\\' {
		// Fast path. No escaped ".
		return data[:n], data[n+1:], nil
	}

	// Slow path - possible escaped " found.
	ss := data
	for {
		i := n - 1
		for i > 0 && data[i-1] == '\\' {
			i--
		}
		if uint(n-i)%2 == 0 {
			return ss[:len(ss)-len(data)+n], data[n+1:], nil
		}
		data = data[n+1:]

		n = strings.IndexByte(data, '"')
		if n < 0 {
			return ss, "", fmt.Errorf(`missing closing '"'`)
		}
		if n == 0 || data[n-1] != '\\' {
			return ss[:len(ss)-len(data)+n], data[n+1:], nil
		}
	}
}

// validateKey is similar to validateString, but is optimized
// for typical object keys, which are quite small and have no escape sequences.
func (json *JsonParser) validateKey(data string) (string, error) {
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			// Fast path - the key doesn't contain escape sequences.
			return data[i+1:], nil
		}
		if data[i] == '\\' {
			// Slow path - the key contains escape sequences.
			return json.parseString(data)
		}
	}
	return data, fmt.Errorf(`missing closing '"'`)
}

func (json *JsonParser) byteArrayToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
