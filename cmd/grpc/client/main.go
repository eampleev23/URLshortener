package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "grpc/proto"
	"log"
)

func main() {
	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// получаем переменную интерфейсного типа ShortenerClient,
	// через которую будем отправлять сообщения
	c := pb.NewShortenerClient(conn)
	// функция, в которой будем отправлять сообщения
	TestShortener(c)
}

func TestShortener(c pb.ShortenerClient) {
	// набор тестовых данных
	urls := []*pb.OriginalURL{
		{OriginalUrl: "http://www.test.com"},
		{OriginalUrl: "http://www.ampleev.com"},
		{OriginalUrl: "http://www.yandex.ru"},
		{OriginalUrl: "http://www.google.ru"},
		{OriginalUrl: "http://www.google.ru"},
	}
	// перебираем все тестовые данные
	for _, original_url := range urls {
		// добавляем урлы
		resp, err := c.AddShortURL(context.Background(), &pb.AddShortURLRequest{OriginalUrl: original_url})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(resp.ShortUrl)
	}
}
