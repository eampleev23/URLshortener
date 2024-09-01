package generatelinks

import "testing"

/*
Для запуска замеров необходимо запустить команду go test -bench . в текущей директории

Результаты для моей среды при генерации 5-символьной короткой ссылки:
BenchmarkGenerateShortURL-14            17_465_426                68.14 ns/op
*/

// BenchmarkGenerateShortURL тетсирует сколько раз успеет отработать функция GenerateShortURL по умолчанию за 1 секунду
func BenchmarkGenerateShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateShortURL()
	}
}
