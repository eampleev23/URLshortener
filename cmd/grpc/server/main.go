package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"

	// импортируем пакет со сгенерированными protobuf-файлами
	pb "grpc/proto"
)

// ShortenerServer поддерживает все необходимые методы сервера.
type ShortenerServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortenerServer
	// используем мапу как в мемори сторе для хранения сокращенных ссылок.
	URLs map[string]string
}

func (s *ShortenerServer) AddShortURL(ctx context.Context, in *pb.AddShortURLRequest) (*pb.AddShortURLResponse, error) {
	// заводим переменную типа ответа
	var response pb.AddShortURLResponse
	// проверяем есть ли такой оригинальный урл в базе и если есть
	if _, ok := s.URLs[in.OriginalUrl.OriginalUrl]; ok {
		// то возвращаем ранее сохраненный сокращенныйц вариант
		response.ShortUrl = fmt.Sprintf("%s", s.URLs[in.OriginalUrl.OriginalUrl])
	} else {
		//s.URLs[in.OriginalUrl.OriginalUrl] = "заглушка сокращенной ссылки"
		s.URLs[in.OriginalUrl.OriginalUrl] = GenerateShortURL()
		response.ShortUrl = fmt.Sprintf("%s", s.URLs[in.OriginalUrl.OriginalUrl])
	}
	return &response, nil
}

func main() {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer()
	// регистрируем сервис
	pb.RegisterShortenerServer(s, &ShortenerServer{
		URLs: make(map[string]string),
	})
	fmt.Println("Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}

// GenerateShortURL Вспомогательная функция для генерации коротких ссылок.
func GenerateShortURL() string {
	// заводим слайс рун возможных для сгенерированной короткой ссылки
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	lenLetterRunes := len(letterRunes)
	// задаем количество символов в сгенерированной ссылке
	numberOfSimbols := 5
	b := make([]rune, numberOfSimbols)

	// генерируем случайный символ последовательно для всей длины
	for i := range b {
		b[i] = letterRunes[rand.Intn(lenLetterRunes)]
	}
	// в результат записываем байты преобразованные в строку
	strResult := string(b)
	return strResult
}
