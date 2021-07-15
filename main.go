package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reggen/reggen"
	"time"

	"github.com/flier/gohs/hyperscan"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func eventHandler(id uint, from, to uint64, flags uint, context interface{}) error {
	return nil
}

func match(regex, inputString string) time.Duration {
	pattern := hyperscan.NewPattern(regex, hyperscan.DotAll|hyperscan.AllowEmpty)
	input := []byte(inputString)

	database, err := hyperscan.NewBlockDatabase(pattern)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	scratch, err := hyperscan.NewScratch(database)
	if err != nil {
		panic(err)
	}
	defer scratch.Free()

	start := time.Now()
	if err := database.Scan(input, scratch, eventHandler, nil); err != nil {
		panic(err)
	}
	duration := time.Since(start)

	return duration
}

func main() {
	file, err := os.Open("regex.txt")
	if err != nil {
		panic(err)
	}
	regexs := []string{}
	ruleId := []string{}
	count := 0
	reader := bufio.NewReaderSize(file, 10000)
	for line, _, err := reader.ReadLine(); err != io.EOF; line, _, err = reader.ReadLine() {
		if count%2 == 0 {
			ruleId = append(ruleId, string(line))
		} else {
			regexs = append(regexs, string(line))
		}
		count++
	}
	regex_res := make([]time.Duration, len(regexs))
	random_res := make([]time.Duration, len(regexs))

	order := make([]int, 20*len(regexs))
	for i := 0; i < len(order); i++ {
		order[i] = i / 20
	}

	cnt := 0
	for _, index := range order {
		gen_string, _ := reggen.Generate(regexs[index], 100)
		random_string := RandStringRunes(len(gen_string))
		regex_res[index] += match(regexs[index], gen_string)
		random_res[index] += match(regexs[index], random_string)
		cnt++
		if cnt == 20 {
			fmt.Printf(
				"[ \033[1;37;42m%s\033[0m | \033[1;33;40m%s\033[0m | \033[1;37;44m%d µs\033[0m | \033[1;37;46m%d µs\033[0m | (%d/%d) ]\n",
				ruleId[index],
				regexs[index],
				regex_res[index].Microseconds(),
				random_res[index].Microseconds(),
				index+1,
				len(regexs),
			)
			cnt = 0
		}
	}
	fmt.Println("Completed!")
}
