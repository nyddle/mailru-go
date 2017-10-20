package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type goCounter struct {
	counts int
	source string
	error error
}

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
		return  "", fmt.Errorf("Unknown source type: %s", sourceType)
	}
}

func getIt(sourceType string, address string, collector chan goCounter) {
	data, err := fetchData(sourceType, address)
	if err != nil {
		collector <- goCounter{error:err}
	} else {
		collector <- goCounter{source: address, counts: countGo(data)}
	}
}

func main() {

	sourceType := flag.String("type", "", "-url http://ya.ru/")
	flag.Parse()

	counts := make(chan goCounter, 5)

	lines := 0
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		address := scanner.Text()
		lines++
		go getIt(*sourceType, address, counts)
	}

	total := 0
	for datum := range counts {
		if datum.error != nil {
			fmt.Printf("Error getting data for %s\n", datum.source)
		} else {
			fmt.Printf("Count for %s: %d\n", datum.source, datum.counts)
			total += datum.counts
		}
		lines--
		if lines == 0 {
			close(counts)
		}
	}
	fmt.Printf("Total: %d\n", total)
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
