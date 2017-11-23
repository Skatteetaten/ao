package collections

func NewStringSet() *StringSet {
	return &StringSet{
		set: make(map[string]bool),
	}
}

type StringSet struct {
	set map[string]bool
}

func (s *StringSet) Add(value string) {
	_, found := s.set[value]
	if !found {
		s.set[value] = true
	}
}

func (s *StringSet) All() []string {
	all := []string{}
	for k, _ := range s.set {
		all = append(all, k)
	}

	return all
}
