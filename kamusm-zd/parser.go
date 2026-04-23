package kamusmzd

import (
	"bytes"
	"regexp"
	"strconv"
	"unicode"
)

var pkcs7SignedDataOID = []byte{0x06, 0x09, 0x2A, 0x86, 0x48, 0x86, 0xF7, 0x0D, 0x01, 0x07, 0x02}

// IsValidTimestampResponse checks if the response body contains a PKCS#7 SignedData OID.
func IsValidTimestampResponse(body []byte) bool {
	return bytes.Contains(body, pkcs7SignedDataOID)
}

// ExtractPkcs7 extracts the PKCS#7 SignedData structure from the response body.
// It searches backward from the OID position to find the enclosing SEQUENCE tag.
func ExtractPkcs7(buf []byte) []byte {
	pos := bytes.Index(buf, pkcs7SignedDataOID)
	if pos < 0 {
		return nil
	}

	startSearch := pos - 16
	if startSearch < 0 {
		startSearch = 0
	}

	for i := pos; i >= startSearch; i-- {
		if buf[i] != 0x30 {
			continue
		}

		if i+1 >= len(buf) {
			continue
		}

		lenByte := buf[i+1]
		var totalLen int

		if lenByte&0x80 == 0 {
			totalLen = int(lenByte) + 2
		} else {
			numBytes := int(lenByte & 0x7F)
			if numBytes == 0 || i+1+numBytes >= len(buf) {
				continue
			}
			if numBytes > 4 {
				continue
			}
			l := 0
			for _, b := range buf[i+2 : i+2+numBytes] {
				l = (l << 8) | int(b)
			}
			totalLen = l + 2 + numBytes
		}

		if i+totalLen <= len(buf) && pos < i+totalLen {
			return buf[i : i+totalLen]
		}
	}

	return nil
}

// ExtractTextFromAsn1 scans the body for ASN.1 string types and extracts printable text.
func ExtractTextFromAsn1(body []byte) []string {
	var texts []string
	i := 0

	for i < len(body)-2 {
		tag := body[i]
		length := body[i+1]

		switch tag {
		case 0x0C, 0x13, 0x14, 0x16, 0x19, 0x1A, 0x1B, 0x1C:
			if length > 0 && i+2+int(length) <= len(body) {
				textBytes := body[i+2 : i+2+int(length)]
				text := string(textBytes)
				trimmed := stringTrimSpace(text)
				if len(trimmed) > 0 && isAsciiPrintable(trimmed) {
					texts = append(texts, trimmed)
				}
				i += 2 + int(length)
			} else {
				i++
			}
		default:
			i++
		}
	}

	return texts
}

func isAsciiPrintable(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
		if !unicode.IsPrint(c) && !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func stringTrimSpace(s string) string {
	result := make([]byte, 0, len(s))
	for _, b := range []byte(s) {
		if b >= 0x20 && b <= 0x7E || b == '\t' || b == '\n' || b == '\r' {
			result = append(result, b)
		}
	}
	trimmed := string(result)
	start := 0
	for start < len(trimmed) && (trimmed[start] == ' ' || trimmed[start] == '\t' || trimmed[start] == '\n' || trimmed[start] == '\r') {
		start++
	}
	end := len(trimmed)
	for end > start && (trimmed[end-1] == ' ' || trimmed[end-1] == '\t' || trimmed[end-1] == '\n' || trimmed[end-1] == '\r') {
		end--
	}
	return trimmed[start:end]
}

var creditRegex = regexp.MustCompile(`(\d+)`)

// ParseCreditsFromBody extracts the first number from the response body.
func ParseCreditsFromBody(body []byte) (uint32, bool) {
	match := creditRegex.Find(body)
	if match == nil {
		return 0, false
	}

	n, err := strconv.ParseUint(string(match), 10, 32)
	if err != nil {
		return 0, false
	}

	return uint32(n), true
}
