package adb

import "bytes"

func Gt(key string, val string) []byte {
	return C(key, GT, val)
}

func Lt(key string, val string) []byte {
	return C(key, LT, val)
}

func Eq(key string, val string) []byte {
	return C(key, EQ, val)
}

func Ne(key string, val string) []byte {
	return C(key, NE, val)
}

func gt(b1, b2 []byte) bool {
	if len(b1) > len(b2) {
		return true
	}
	if len(b1) == len(b2) {
		return bytes.Compare(b1, b2) == 1
	}
	return false
}

func lt(b1, b2 []byte) bool {
	if bytes.Equal(b1, b2) {
		return false
	}
	return !gt(b1, b2)
}

const (
	LT byte = iota
	GT
	EQ
	NE
)

func C(key string, op byte, val string) []byte {
	b := [][]byte{[]byte(`"` + key + `":`)}
	b = append(b, []byte{op})
	if val[0] == '"' && val[len(val)-1] == '"' {
		return bytes.Join(append(b, []byte(val)), []byte{'_'})
	}
	if val == "true" || val == "false" {
		return bytes.Join(append(b, []byte(val)), []byte{'_'})
	}
	isDigit := true
	for _, c := range val {
		if c < '0' || c > '9' {
			isDigit = false
			break
		}
	}
	if isDigit {
		return bytes.Join(append(b, []byte(val)), []byte{'_'})
	}
	return bytes.Join(append(b, []byte(`"`+val+`"`)), []byte{'_'})

}

func check(q, d []byte) bool {
	qry := bytes.Split(q, []byte{'_'})
	key := qry[0]
	val := qry[2]
	innerVal, ok := getValByKey(key, d)
	if !ok {
		return ok
	}
	switch qry[1][0] {
	case GT:
		return gt(innerVal, val)
	case LT:
		return lt(innerVal, val)
	case EQ:
		return bytes.Equal(innerVal, val)
	case NE:
		return !bytes.Equal(innerVal, val)
	}
	return false
}

// expects key to be in following format: `"foo":`...
// ie. starting with a quote and ending in a quote then colon.
func getValByKey(k, d []byte) ([]byte, bool) {
	idx := bytes.Index(d, k)
	if idx == -1 {
		return nil, false
	}
	offset := idx + len(k)
	if d[offset] == '"' {
		idx = bytes.Index(d[offset:], []byte{'"', ','})
		if idx == -1 {
			idx = bytes.Index(d[offset:], []byte{'"', '}'})
			if idx == -1 {
				return nil, false
			}
		}
		return d[offset : offset+idx+1], true
	}
	idx = bytes.Index(d[offset:], []byte{','})
	if idx == -1 {
		idx = bytes.Index(d[offset:], []byte{'}'})
		if idx == -1 {
			return nil, false
		}
	}
	return d[offset : offset+idx], true
}
