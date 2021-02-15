package tools

func Strcasecmp(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		c1 := s1[i]
		c2 := s2[i]

		if c1 != c2 {
			if 'A' <= c1 && c1 <= 'Z' {
				c1 += 'a' - 'A'
			}

			if 'A' <= c2 && c2 <= 'Z' {
				c2 += 'a' - 'A'
			}

			if c1 != c2 {
				return false
			}
		}
	}

	return true
}
