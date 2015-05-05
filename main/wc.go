package main

import (
	"container/list"
	"fmt"
	"github.com/keathley/6.824/mapreduce"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// our simplified version of MapReduce does not supply a
// key to the Map function, as in the paper; only a value,
// which is a part of the input file contents
func Map(value string) *list.List {
	words := list.New()
	split := strings.FieldsFunc(value, isNonWord)
	for i := 0; i < len(split); i++ {
		words.PushBack(mapreduce.KeyValue{split[i], "1"})
	}
	return words
}

func isNonWord(c rune) bool {
	return !unicode.IsLetter(c)
}

// iterate over list and add values
func Reduce(key string, values *list.List) string {
	sum := 0
	for e := values.Front(); e != nil; e = e.Next() {
		i, err := strconv.ParseInt(e.Value.(string), 10, 0)
		if err == nil {
			sum += int(i)
		}
	}
	return strconv.Itoa(sum)
}

// Can be run in 3 ways:
// 1) Sequential (e.g., go run wc.go master x.txt sequential)
// 2) Master (e.g., go run wc.go master x.txt localhost:7777)
// 3) Worker (e.g., go run wc.go worker localhost:7777 localhost:7778 &)
func main() {
	if len(os.Args) != 4 {
		fmt.Printf("%s: see usage comments in file\n", os.Args[0])
	} else if os.Args[1] == "master" {
		if os.Args[3] == "sequential" {
			mapreduce.RunSingle(5, 3, os.Args[2], Map, Reduce)
		} else {
			mr := mapreduce.MakeMapReduce(5, 3, os.Args[2], os.Args[3])
			// Wait until MR is done
			<-mr.DoneChannel
		}
	} else {
		mapreduce.RunWorker(os.Args[2], os.Args[3], Map, Reduce, 100)
	}
}
