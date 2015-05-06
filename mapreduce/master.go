package mapreduce

import (
	"container/list"
	"fmt"
	"sync"
)

type WorkerInfo struct {
	address string
	// You can add definitions here.
}

// Clean up all workers by sending a Shutdown RPC to each one of them
// Collect the number of jobs each worker has performed.
func (mr *MapReduce) KillWorkers() *list.List {
	l := list.New()
	for _, w := range mr.Workers {
		DPrintf("DoWork: shutdown %s\n", w.address)
		args := &ShutdownArgs{}
		var reply ShutdownReply
		ok := call(w.address, "Worker.Shutdown", args, &reply)
		if ok == false {
			fmt.Printf("DoWork: RPC %s shutdown error\n", w.address)
		} else {
			l.PushBack(reply.Njobs)
		}
	}
	return l
}

func (mr *MapReduce) RunMaster() *list.List {
	// Your code here

	mapChan := make(chan int, mr.nMap)
	mappedChan := make(chan int, mr.nMap)

	for maps := 0; maps < mr.nMap; maps++ {
		go func(jobNum int) {
			worker := GetNextWorker(mr)
			mapChan <- 1
			repl := &DoJobReply{}
			SendJob(mr, worker, Map, jobNum, mr.nReduce, repl)
			mappedChan <- 1
		}(maps)
	}

	// reduce jobs
	var wg sync.WaitGroup
	for reduces := 0; reduces < mr.nReduce; reduces++ {
		wg.Add(1)
		go func(jobNum int) {
			defer wg.Done()
			worker := GetNextWorker(mr)
			repl := &DoJobReply{}
			SendJob(mr, worker, Reduce, jobNum, mr.nMap, repl)
		}(reduces)
	}

	wg.Wait()
	return mr.KillWorkers()
}

func GetNextWorker(mr *MapReduce) string {
	return <-mr.registerChannel
}

func SendJob(mr *MapReduce, worker string, op JobType, jobNum int, others int, repl *DoJobReply) {
	args := &DoJobArgs{
		File:          mr.file,
		Operation:     op,
		JobNumber:     jobNum,
		NumOtherPhase: others,
	}
	call(worker, "Worker.DoJob", args, repl)
	select {
	case mr.registerChannel <- worker:
		// add it and move on
	default:
		// just return
	}
}
