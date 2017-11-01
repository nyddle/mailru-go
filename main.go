package main

import (
	"bufio"
	"flag"
	"sync"
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
	for i := 0; i < len(input)-1; i++ {
		if input[i] == 'G' && input[i+1] == 'o' {
			count++
		}
	}
	return count
}

func getURL(url string) (string, error) {

	resp, err := http.Get(url)
	if err != nil {
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

func getIt(sourceType string, address string, collector chan goCounter, wg *sync.WaitGroup) {
	data, err := fetchData(sourceType, address)
	if err != nil {
		collector <- goCounter{error:err}
	} else {
		collector <- goCounter{source: address, counts: countGo(data)}
	}
	wg.Done()
}

func main() {

	sourceType := flag.String("type", "", "-url http://ya.ru/")
	flag.Parse()

	counts := make(chan goCounter, 5)

	scanner := bufio.NewScanner(os.Stdin)
	var wg sync.WaitGroup
	for scanner.Scan() {
		wg.Add(1)
		address := scanner.Text()
		go getIt(*sourceType, address, counts, &wg)
	}

	wg.Wait()
	close(counts)

	total := 0
	for datum := range counts {
		if datum.error != nil {
			fmt.Printf("Error getting data for %s: %s\n", datum.source, datum.error.Error())
		} else {
			fmt.Printf("Count for %s: %d\n", datum.source, datum.counts)
			total += datum.counts
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
*/
