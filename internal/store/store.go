package store

type Store struct {
	s map[string]string
}

func NewStore() *Store {
	return &Store{
		s: make(map[string]string),
	}
}

func (s *Store) SetShortURL(shortURL string, longURL string) error {
	s.s[shortURL] = longURL
	return nil
}
