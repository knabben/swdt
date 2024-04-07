package exec

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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

func EnableOutput(output *string, fn func(std *chan string)) {
	std := make(chan string)
	fn(&std)
	var outlist []string
	for {
		select {
		case n, ok := <-std:
			if !ok {
				if output != nil {
					*output = strings.Join(outlist, " ")
				}
				break
			}
			if output != nil {
				outlist = append(outlist, n)
			}
			fmt.Println(n)
		}
	}
}
