package resp

import "strconv"

func EncodeSimpleString(s string) []byte {
	return []byte("+" + s + "\r\n")
}

func EncodeError(s string) []byte {
	return []byte("-ERR " + s + "\r\n")
}

func EncodeInteger(n int) []byte {
	return []byte(":" + strconv.Itoa(n) + "\r\n")
}

func EncodeBulkString(s string) []byte {
	return []byte("$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n")
}

func EncodeNullBulkString() []byte {
	return []byte("$-1\r\n")
}

func EncodeArray(elems [][]byte) []byte {
	result := []byte("*" + strconv.Itoa(len(elems)) + "\r\n")
	for _, e := range elems {
		result = append(result, e...)
	}
	return result
}
