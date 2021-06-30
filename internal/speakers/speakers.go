package speakers

import (
	"bufio"
	"log"
	"os"
)

var speakers map[string]int

func init() {
	file, err := os.Open("/speakers")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	i := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		speakers[scanner.Text()] = i
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func Lookup(s string) int {
	i, _ := speakers[s]
	return i
}
