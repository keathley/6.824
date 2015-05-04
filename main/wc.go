package main

import "os"
import "fmt"
import "github.com/keathley/6.824/mapreduce"
import "container/list"
import "unicode"
import "strings"
import "strconv"

// Map lists each word in a document.
// Our simplified version of MapReduce does not supply a
// key to the Map function, as in the paper; only a value,
// which is a part of the input file contents.
// key: document name
// value: document contents
// for each word w in value:
//   EmitIntermediate(w, "1");
func Map(value string) *list.List {
	l := list.New()
	// A word is defined as any contiguous sequence of letters
	f := func(c rune) bool { return !unicode.IsLetter(c) }
	for _, w := range strings.FieldsFunc(value, f) {
		l.PushFront(mapreduce.KeyValue{w, "1"})
	}
	return l
}

// Reduce iterates over list and adds values.
// key: a word
// values: a list of counts
// int result = 0;
// for each v in values:
//   result += ParseInt(v);
// Emit(AsString(result));
func Reduce(key string, values *list.List) string {
	r := 0
	for v := values.Front(); v != nil; v = v.Next() {
		n, _ := strconv.ParseInt(v.Value.(string), 0, 0)
		r += int(n)
	}
	return strconv.Itoa(r)
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
