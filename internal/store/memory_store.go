package store

import "context"

type MemoryStore struct {
	ms map[string]LinksCouple
}

func (ms *MemoryStore) SetShortURL(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL = ""
	return shortURL, nil
}
func (ms *MemoryStore) GetOriginalURLByShort(ctx context.Context, shortURL string) (originalURL string, err error) {
	originalURL = ""
	return originalURL, nil
}
func (ms *MemoryStore) GetShortURLByOriginal(ctx context.Context, originalURL string) (shortURL string, err error) {
	shortURL = ""
	return shortURL, nil
}
