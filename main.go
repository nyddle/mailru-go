package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func countGo(input string) int {

	count := 0
	for i := 0; i < len(input); i++ {
		if input[i] == 'g' && input[i+1] == 'o' {
			count++
		}
	}
	return count
}

func getURL(url string) (string, error) {

	resp, err := http.Get(url)
	if err != nil {
		//TODO: warn here
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), err
	}
	return string(body), nil
}

func readFile(filename string) (string, error) {

	data, err := ioutil.ReadFile(filename)
	return string(data), err
}

func fetchData(sourceType string, address string) (string, error) {

	switch sourceType {
	case "url":
		return getURL(address)
	case "file":
		return readFile(address)
	default:
		panic("Unknown option")
	}
}

func getIt(sourceType string, address string, collector chan string) {
	data, err := fetchData(sourceType, address)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ddv")
	}
	collector <- fmt.Sprintf("Count for %s: %d\n",  address, countGo(data))
}

func main() {

	sourceType := flag.String("type", "", "-url http://ya.ru/")
	flag.Parse()

	counts := make(chan string, 5)
	total := make(chan int, 5)


	lines := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		address := scanner.Text()
		lines++
		go getIt(*sourceType, address, counts)
	}

	for msg := range counts {
		fmt.Print(msg)
		lines--
		if lines == 0 {
			close(counts)
		}
	}
}

/*
роцессу на stdin приходят строки, содержащие URL или названия файлов.
Что именно приходит на stdin определяется с помощью параметра командной строки -type. Например, -type file или -type url.
Каждый такой URL нужно запросить, каждый файл нужно прочитать, и посчитать кол-во вхождений строки "Go" в ответе.
В конце работы приложение выводит на экран общее кол-во найденных строк "Go" во всех источниках данных, например:

$ echo -e 'https://golang.org\nhttps://golang.org\nhttps://golang.org' | go run 1.go -type url
Count for https://golang.org: 9
Count for https://golang.org: 9
Count for https://golang.org: 9
Total: 27
$ echo -e '/etc/passwd\n/etc/hosts' | go run 1.go - type file
Count for /etc/passwd: 0
Count for /etc/hosts: 0
Total: 0
Каждый источник данных должен начать обрабатываться сразу после вычитывания и параллельно с вычитыванием следующего. Источники должны обрабатываться параллельно, но не более k=5 одновременно. Обработчики данных не должны порождать лишних горутин, т.е. если k=1000 а обрабатываемых источников нет, не должно создаваться 1000 горутин.
Нужно обойтись без глобальных переменных и использовать только стандартные библиотеки. Код должен быть написан так, чтобы его можно было легко тестировать.
Формат предоставления решения: ссылка на github.
С уважением,
Мария Тимченко
Менеджер отдела подбора персонала
Mail.Ru Group
Офис.: +7-495-725-63-57 доб. 2281
Моб.: +7-916-157-86-81
Агент: m.timchenko@corp.mail.ru
www.corp.mail.ru
*/
