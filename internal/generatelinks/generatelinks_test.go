package generatelinks

import (
	"fmt"
	"testing"
	"time"
)

// Для запуска замеров необходимо запустить команду "go test -bench ." в текущей директории.
// Результаты для моей среды при генерации 5-символьной короткой ссылки:
// "BenchmarkGenerateShortURL-14            17_465_426                68.14 ns/op".

// BenchmarkGenerateShortURL тетсирует сколько раз успеет отработать функция GenerateShortURL по умолчанию за 1 секунду.
func BenchmarkGenerateShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateShortURL()
	}
}

// ExampleGenerateShortURL - пример использования функции генерации короткой ссылки
// и одновременно тест проверки допустимого времени генерации.
func ExampleGenerateShortURL() {
	start := time.Now()
	// запускаем функцию
	_ = GenerateShortURL()
	// измеряем время генерации
	duration := time.Since(start)
	if duration > time.Millisecond {
		fmt.Printf("Duration GenerateShortURL > time.Second, the value is %s (OMG!!!)\n", duration)
	} else {
		fmt.Println("Ok")
	}
	// Output:
	// Ok
}
