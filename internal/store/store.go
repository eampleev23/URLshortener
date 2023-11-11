package store

type Store struct {
	s map[string]string
}

func NewStore() *Store {
	return &Store{
		s: make(map[string]string),
	}
}

func (s *Store) SetShortUrl(shortUrl string, longUrl string) error {
	s.s[shortUrl] = longUrl
	return nil
}
