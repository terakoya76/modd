package filter

// Intersect returns the product set of arguments
func Intersect(l1, l2 []string) []string {
	s := make(map[string]struct{}, len(l1))
	for _, elmt := range l1 {
		s[elmt] = struct{}{}
	}

	r := make([]string, 0, len(l1))
	for _, elmt := range l2 {
		if _, ok := s[elmt]; !ok {
			continue
		}

		r = append(r, elmt)
	}

	return r
}

// Difference returns the difference set of arguments
func Difference(l1, l2 []string) []string {
	s := make(map[string]struct{}, len(l1))
	for _, elmt := range l2 {
		s[elmt] = struct{}{}
	}

	r := make([]string, 0, len(l2))
	for _, elmt := range l1 {
		if _, ok := s[elmt]; ok {
			continue
		}

		r = append(r, elmt)
	}

	return r
}
