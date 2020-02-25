package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// The project version
const version = "1.0.0"

// The max count of the goroutines which work simultaneously
const maxConcurrency = 5

// Look for the following word
const target = "Go"

type semaphoreInterface interface {
	Acquire()
	Release()
}

type semaphoreImpl struct {
	sem chan struct{}
}

func (s *semaphoreImpl) Acquire() {
	s.sem <- struct{}{}
}

func (s *semaphoreImpl) Release() {
	_ = <-s.sem
}

func semaphore(tickets int) semaphoreInterface {
	return &semaphoreImpl{
		sem: make(chan struct{}, tickets),
	}
}

type resultWriter = func(string, int)

func urlHandler(url string, resWriter resultWriter) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "urlHandler:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "urlHandler:", err)
		return
	}

	resWriter(url, strings.Count(string(body), target))
}

func stdInProcessing(resWriter resultWriter) {

	wg := new(sync.WaitGroup)
	s := semaphore(maxConcurrency)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "reading standard input error: %s. Ignoring line", err)
			continue
		}

		rawURL := scanner.Text()
		if rawURL == "" {
			continue
		}
		if rawURL == "exit" {
			break
		}

		_, err := url.Parse(rawURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "It isn't a valid url: %s. Skipping...", rawURL)
			continue
		}

		s.Acquire()

		wg.Add(1)
		go func() {
			urlHandler(rawURL, resWriter)
			s.Release()
			wg.Done()
		}()
	}

	wg.Wait()
}

func main() {
	fmt.Printf("FinalTask v%s, Feb 2020\n", version)
	fmt.Println("Please, input the URL, URL list or 'exit'")

	mutex := new(sync.Mutex)
	var total int64

	resWriter := func(url string, n int) {
		mutex.Lock()
		defer mutex.Unlock()
		fmt.Printf("Count for %s: %d\n", url, n)
		total += int64(n)
	}

	stdInProcessing(resWriter)

	fmt.Println("Total:", total)
	fmt.Println("Successful complete")
}
