package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func (px *Proxy) Handle(w http.ResponseWriter, r *http.Request) {
	parsedURL, err := url.Parse(r.RequestURI)
	if err != nil {
		http.Error(w, "Некорректный URL", http.StatusBadRequest)
		log.Println(err.Error())
		return
	}

	targetHost := parsedURL.Host
	if targetHost == "" {
		targetHost = r.Host
	}
	if !strings.Contains(targetHost, ":") {
		targetHost += ":80"
	}

	// Устанавливаем TCP-соединение с целевым сервером
	connection, err := net.Dial("tcp", targetHost)
	if err != nil {
		http.Error(w, "Ошибка подключения к хосту", http.StatusBadGateway)
		return
	}
	defer connection.Close()

	// Формируем и отправляем HTTP-запрос
	requestLine := fmt.Sprintf("%s %s %s\r\n", r.Method, parsedURL.RequestURI(), r.Proto)
	connection.Write([]byte(requestLine))

	for header, values := range r.Header {
		if strings.EqualFold(header, "Proxy-Connection") {
			continue
		}
		for _, value := range values {
			connection.Write([]byte(fmt.Sprintf("%s: %s\r\n", header, value)))
		}
	}
	connection.Write([]byte(fmt.Sprintf("Host: %s\r\n", targetHost)))
	connection.Write([]byte("\r\n"))

	if r.Method == "POST" || r.Method == "PUT" {
		io.Copy(connection, r.Body) // Передаем тело запроса на сервер
	}

	// Читаем ответ от сервера и отправляем его клиенту
	responseReader := bufio.NewReader(connection)
	response, err := http.ReadResponse(responseReader, r)
	if err != nil {
		http.Error(w, "Ошибка при чтении ответа", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	io.Copy(w, response.Body) // Отправляем тело ответа клиенту
}
