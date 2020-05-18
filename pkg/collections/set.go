package collections

// NewStringSet makes a new StringSet
func NewStringSet() *StringSet {
	return &StringSet{
		set: make(map[string]bool),
	}
}

// StringSet is a set of strings
type StringSet struct {
	set map[string]bool
}

// Add adds a value to a StringSet
func (s *StringSet) Add(value string) {
	_, found := s.set[value]
	if !found {
		s.set[value] = true
	}
}

// All returns an array of all strings in a StringSet
func (s *StringSet) All() []string {
	all := []string{}
	for k := range s.set {
		all = append(all, k)
	}

	return all
}
