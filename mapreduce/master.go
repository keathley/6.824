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

	for maps := 0; maps < mr.nMap; maps++ {
		go func(jobNum int) {
			WorkUntilComplete(mr, Map, jobNum, mr.nReduce, done)
		}(maps)
	}

	// reduce jobs
	for reduces := 0; reduces < mr.nReduce; reduces++ {
		go func(jobNum int) {
			WorkUntilComplete(mr, Reduce, jobNum, mr.nMap, done)
		}(reduces)
	}

	<-done
	return mr.KillWorkers()
}

func WorkUntilComplete(mr *MapReduce, op JobType, jobNum int, others int, done chan int) {
	reply := &DoJobReply{}
	worker := GetNextWorker(mr)
	fmt.Printf("Sending job %d\n", jobNum)
	rpcResp := SendJob(mr, worker, op, jobNum, others, reply, done)
	if !rpcResp || !reply.OK {
		WorkUntilComplete(mr, op, jobNum, others, done)
	}
}

func GetNextWorker(mr *MapReduce) string {
	return <-mr.registerChannel
}

func SendJob(mr *MapReduce, worker string, op JobType, jobNum int, others int, repl *DoJobReply, done chan int) bool {
	args := &DoJobArgs{
		File:          mr.file,
		Operation:     op,
		JobNumber:     jobNum,
		NumOtherPhase: others,
	}
	resp := call(worker, "Worker.DoJob", args, repl)
	if resp && repl.OK {
		select {
		case mr.registerChannel <- worker:
			// re-register it and move on
		default:
			// no more workers being requested: we're done
			done <- 1
		}
	} else {
		fmt.Printf("FAILURE: job %d\n", jobNum)
	}
	return resp
}
