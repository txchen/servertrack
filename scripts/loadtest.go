package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jawher/mow.cli"
)

func main() {
	app := cli.App("loadtest", "add data to servertrack")
	servername := app.StringOpt("s serverName", "", "server name")
	count := app.IntOpt("c count", 10000, "call count")
	thread := app.IntOpt("t thread", 3, "thread number")
	dataDate := app.StringOpt("d date", "", "data date e.g. 20150406")
	app.Spec = "-s -d [-c] [-t] "
	app.Action = func() {
		parsedDate, err := time.Parse("20060102", *dataDate)
		if err != nil {
			app.PrintHelp()
			return
		}

		start := time.Now()
		tasks := make(chan int64, 64)
		// spawn worker goroutines
		var wg sync.WaitGroup
		for i := 0; i < *thread; i++ {
			wg.Add(1)
			go func() {
				client := &http.Client{}
				for unixTime := range tasks {
					url := fmt.Sprintf("http://localhost:30000/load?servername=%s&cpuload=%f&ramload=%f&unixtime=%d",
						*servername, rand.Float32()*100, rand.Float32()*100, unixTime)
					//fmt.Println(url)
					req, _ := http.NewRequest("POST", url, nil)
					resp, err := client.Do(req)
					if err != nil {
						panic(err)
					}
					resp.Body.Close()
				}
				wg.Done()
			}()
		}
		// generate tasks
		for i := 0; i < *count; i++ {
			tasks <- parsedDate.Unix() + rand.Int63n(86400)
		}
		close(tasks)

		// wait for the workers to finish
		wg.Wait()
		fmt.Printf("Finished sent %d records with concurrency %d\n", *count, *thread)
		fmt.Printf("Spent %s\n", time.Since(start))
	}
	app.Run(os.Args)
}
