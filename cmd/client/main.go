package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func clientGenerateUrl() {
	// ---- Клиент для тестирования генерации короткого URL
	endpoint := "http://localhost:8080/"
	// контейнер данных для запроса
	data := url.Values{}
	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")
	// заполняем контейнер данными
	data.Set("url", long)
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "text/plain")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	// и печатаем его
	fmt.Println(string(body))
}

func clientUseUrl() {
	// ---- Клиент для тестирования генерации короткого URL
	endpoint := "http://localhost:8080/"
	//контейнер данных для запроса
	//приглашение в консоли
	fmt.Println("Введите URL для проверки")
	//открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	//читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")
	//подставляем запрос к эндпоинту
	endpoint = endpoint + strings.TrimSuffix(long, "\n")
	// добавляем HTTP-клиент
	client := &http.Client{}
	request, err := client.Get(endpoint)
	if err != nil {
		panic(err)
	}
	fmt.Println("Статус-код ", request.Status)

}

func main() {
	clientGenerateUrl()
	//clientUseUrl()
}
