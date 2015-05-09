package mapreduce

import (
	"container/list"
	"fmt"
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
	done := make(chan int)

	mapChan := make(chan int, mr.nMap)
	mappedChan := make(chan int, mr.nMap)
	fillBuffer(mapChan, mr.nMap)
	for m := range mapChan {
		go func(jobNum int) {
			WorkUntilComplete(mr, Map, jobNum, mr.nMap, mr.nReduce, mapChan, mappedChan, done)
		}(m)
	}
	<-done
	DPrintf("Value received from done channel\n.")
	reduceChan := make(chan int, mr.nReduce)
	reducedChan := make(chan int, mr.nReduce)
	fillBuffer(reduceChan, mr.nReduce)
	for r := range reduceChan {
		go func(jobNum int) {
			WorkUntilComplete(mr, Reduce, jobNum, mr.nReduce, mr.nMap, reduceChan, reducedChan, done)
		}(r)
	}

	<-done
	return mr.KillWorkers()
}

func fillBuffer(buff chan int, num int) {
	for i := 0; i < num; i++ {
		buff <- i
	}
}

func WorkUntilComplete(mr *MapReduce, op JobType, jobNum int, lastJob int, others int, jobChan chan int, completedChan chan int, done chan int) {
	reply := &DoJobReply{}
	worker := GetNextWorker(mr)
	DPrintf("Sending %s job %d\n", op, jobNum)
	rpcResp := SendJob(mr, worker, op, jobNum, others, reply, jobChan, done)
	if rpcResp && reply.OK {
		completedChan <- jobNum
		DPrintf("%s job %d/%d completed successfully...channel now has %d jobs\n", op, jobNum, lastJob, len(completedChan))
		if len(completedChan) == lastJob {
			DPrintf("Closing job channel \n")
			close(jobChan)
			done <- 1
		}
	} else {
		WorkUntilComplete(mr, op, jobNum, lastJob, others, jobChan, completedChan, done)
	}
}

func GetNextWorker(mr *MapReduce) string {
	return <-mr.registerChannel
}

func SendJob(mr *MapReduce, worker string, op JobType, jobNum int, others int, repl *DoJobReply, jobChan chan int, done chan int) bool {
	args := &DoJobArgs{
		File:          mr.file,
		Operation:     op,
		JobNumber:     jobNum,
		NumOtherPhase: others,
	}
	resp := call(worker, "Worker.DoJob", args, repl)
	if resp && repl.OK {
		// XXX I'm not sure how much this affects concurrency - in theory, the workers *could* just fail for a single job and work for subsequent ones. I'm treating a failure as a complete failure and not allowing the worker to re-register
		mr.registerChannel <- worker
	} else {
		fmt.Printf("FAILURE: job %d\n", jobNum)
	}
	return resp
}
