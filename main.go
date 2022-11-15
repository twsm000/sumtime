package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var (
	fileTypes map[string]DurationFileReader = map[string]DurationFileReader{
		"txt":  ReadTextFile,
		"json": ReadJSONFile,
	}

	fileType string
)

const (
	appUsage string = `
		Usage: sumtime [-ft] <fn1 fn2 fn3...>
		-ft: file type supported = txt, json. txt is the default
	`
)

func main() {
	flag.StringVar(&fileType, "ft", "txt", "file type")
	flag.Parse()

	readFile, ok := fileTypes[fileType]
	if !ok {
		fmt.Println(appUsage)
		os.Exit(1)
	}

	for summary := range SumDuration(ReadFiles(readFile)) {
		fmt.Println("Summary:", summary.Summarize())
	}
}

func ReadFiles(readFile DurationFileReader) <-chan DurationFileResponse {
	send := make(chan DurationFileResponse)

	go func() {
		defer close(send)
		for _, fileName := range flag.Args() {
			resp, err := readFile(fileName)
			if err != nil {
				fmt.Println(err)
				continue
			}
			send <- resp
		}
	}()

	return send
}

func SumDuration(responseStream <-chan DurationFileResponse) <-chan DurationSummary {
	send := make(chan DurationSummary)

	go func() {
		defer close(send)

		for resp := range responseStream {
			var sum time.Duration
			fileName, durations := resp()
			for _, hms := range durations.Values() {
				d, err := time.ParseDuration(hms)
				if err != nil {
					fmt.Println(`Wrong time format. Accept 00h, 00m or 00s that can be composed. 
						Error:`, err)
					continue
				}
				sum += d
			}
			send <- DurationSummary{fileName: fileName, sum: sum}
		}
	}()

	return send
}

type DurationFileResponse func() (fileName string, ds *Durations)
type DurationFileReader func(filename string) (DurationFileResponse, error)

type Slice[T any] []T
type Strings Slice[string]

type Durations struct {
	Times Strings `json:"times"`
}

func (d *Durations) Add(time string) {
	d.Times = append(d.Times, time)
}

func (d *Durations) Values() Strings {
	return d.Times
}

type DurationSummary struct {
	fileName string
	sum      time.Duration
}

func (ds DurationSummary) Time() (t time.Time) {
	return t.Add(ds.sum)
}

func (ds DurationSummary) Summarize() string {
	return fmt.Sprintf("%s = %s (%s)", ds.fileName, ds.Simplified(), ds.Detailed())
}

func (ds DurationSummary) Simplified() string {
	remain := ds.sum
	h := remain / time.Hour
	remain -= h * time.Hour

	m := remain / time.Minute
	remain -= m * time.Minute

	s := remain / time.Second
	return fmt.Sprintf("%dh %dm %ds", int(h), int(m), int(s))
}

func (ds DurationSummary) Clock() string {
	h, m, s := ds.Time().Clock()
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}

func (ds DurationSummary) Detailed() string {
	d := int(ds.sum.Hours() / 24)
	y := ds.Time().Year() - 1
	d -= y * 365
	return fmt.Sprintf("%dY %dd %s", y, d, ds.Clock())
}
