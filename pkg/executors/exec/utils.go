package exec

import (
	"bufio"
	"io"
	"sync"
)

func redirectStandard(mu *sync.Mutex, std io.Reader, to *chan string) {
	if to != nil {
		mu.Lock()
		defer mu.Unlock()
		scanner := bufio.NewScanner(std)
		for scanner.Scan() {
			*to <- scanner.Text()
		}
		close(*to)
		*to = make(chan string)
	}
}
