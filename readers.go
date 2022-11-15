package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func NewFailedOpenFileError(err error) error {
	return fmt.Errorf("failed to open file: %w", err)
}

func NewDurationResponse(fileName string, ds *Durations) DurationFileResponse {
	return func() (string, *Durations) {
		return fileName, ds
	}
}

func ReadTextFile(fileName string) (DurationFileResponse, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, NewFailedOpenFileError(err)
	}
	defer f.Close()

	ds := &Durations{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		ds.Add(sc.Text())
	}

	return NewDurationResponse(fileName, ds), nil
}

func ReadJSONFile(fileName string) (DurationFileResponse, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, NewFailedOpenFileError(err)
	}
	defer f.Close()

	ds := &Durations{}
	err = json.NewDecoder(f).Decode(&ds)
	if err != nil {
		return nil, fmt.Errorf("failed to parse json file: %s. error: %w", fileName, err)
	}

	return NewDurationResponse(fileName, ds), nil
}
