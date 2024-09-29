package proxy

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"syscall"
)

func (px *Proxy) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("HTTP сервер не поддерживает перехват соединений")
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Fatal("Не удалось перехватить HTTP соединение")
	}
	defer clientConn.Close()

	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		log.Fatal("Ошибка при разбиении хоста/порта:", err)
	}

	caCert, caKey, err := LoadKeyPair(host)
	if err != nil {
		log.Fatal("Не удалось загрузить ключевую пару: " + err.Error())
	}

	pemCert, pemKey := CreateCert([]string{host}, caCert, caKey, 240)
	tlsCert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
		log.Fatal("Ошибка при записи статуса клиенту:", err)
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}
	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	connReader := bufio.NewReader(tlsConn)
	request, err := http.ReadRequest(connReader)
	if err == io.EOF {
		log.Fatal(request, err)
	} else if errors.Is(err, syscall.ECONNRESET) {
		log.Print("Соединение сброшено клиентом")
		log.Fatal(err)
	} else if err != nil {
		log.Fatal(err)
	}

	if b, err := httputil.DumpRequest(request, false); err == nil {
		log.Printf("Получен входящий запрос:\n%s\n", string(b))
	}

	copyRequest := *request
	targetURL := GetURL(request.Host)
	targetURL.Path = request.URL.Path
	targetURL.RawQuery = request.URL.RawQuery
	request.URL = targetURL
	request.RequestURI = ""

	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		log.Println("Ошибка при отправке запроса к цели:", err)
		log.Fatal(request, err)
	}

	if b, err := httputil.DumpResponse(response, true); err == nil {
		log.Printf("Ответ от цели:\n%s\n", string(b))
	}
	defer response.Body.Close()

	if err := response.Write(tlsConn); err != nil {
		log.Println("Ошибка при записи ответа обратно:", err)
	}
	copyRequest.URL.Scheme = "https"
	// _, err = px.uc.SaveRequestResponse(&copyRequest, response)
	// if err != nil {
	//     log.Println("Ошибка при сохранении запроса и ответа:", err)
	// }
}

func GetURL(addr string) *url.URL {
	if !strings.HasPrefix(addr, "https") {
		addr = "https://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}
	return u
}
