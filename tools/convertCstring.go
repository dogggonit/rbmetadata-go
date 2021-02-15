package tools

func CString(s []byte) string {
	clen := func(n []byte) int {
		for i := 0; i < len(n); i++ {
			if n[i] == 0 {
				return i
			}
		}
		return len(n)
	}

	return string(s[:clen(s)])
}
