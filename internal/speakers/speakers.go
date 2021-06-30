package speakers

import (
	"bufio"
	"log"
	"os"
)

var speakers map[string]int
var speakerLogins map[int]string
var speakerNames map[int]string

func init() {
	file, err := os.Open("/speakers")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	i := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		login := scanner.Text()
		speakers[login] = i
		speakerLogins[i] = login
		i++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	file2, err := os.Open("/speakernames")
	if err != nil {
		log.Fatal(err)
	}
	defer file2.Close()

	i = 1
	scanner2 := bufio.NewScanner(file2)
	for scanner2.Scan() {
		speakerNames[i] = scanner2.Text()
		i++
	}

	if err := scanner2.Err(); err != nil {
		log.Fatal(err)
	}

}

func Lookup(s string) int {
	i, _ := speakers[s]
	return i
}

func ReverseLookup(i int) (string, string) {
	login, ok := speakerLogins[i]
	if !ok {
		return "", ""
	}
	name, ok := speakerNames[i]
	if !ok {
		return "", ""
	}
	return login, name
}
