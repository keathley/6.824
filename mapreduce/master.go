package mapreduce

import (
	"container/list"
	"fmt"
	"os"
)

type WorkerInfo struct {
	address string
	// You can add definitions here.
}

// Clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
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
	for mapJobId := 0; mapJobId < mr.nMap; mapJobId++ {
		go func(jobId int) {
			mr.SendJob(jobId, Map, mr.nReduce)
		}(mapJobId)
	}

	for reduceJobId := 0; reduceJobId < mr.nReduce; reduceJobId++ {
		mr.wgReduce.Add(1)
		go func(jobId int) {
			defer mr.wgReduce.Done()
			mr.SendReduceJob(jobId)
		}(reduceJobId)
	}

	fmt.Println("Waiting for jobs to finish")
	mr.wgReduce.Wait()
	fmt.Println("Finished")

	return mr.KillWorkers()
}

func (mr *MapReduce) SendMapJob(jobId int) {
	mr.SendJob(jobId, Map, mr.nReduce)
}

func (mr *MapReduce) SendReduceJob(jobId int) {
	mr.SendJob(jobId, Reduce, mr.nMap)
}

func (mr *MapReduce) SendJob(jobId int, operation JobType, otherCount int) {
	worker := <-mr.registerChannel
	args := &DoJobArgs{
		File:          mr.file,
		Operation:     operation,
		JobNumber:     jobId,
		NumOtherPhase: otherCount,
	}
	reply := &DoJobReply{}

	ok := call(worker, "Worker.DoJob", args, reply)

	if ok && reply.OK == true {
		fmt.Println("Worker is done", operation, jobId, reply, ok)
		select {
		case mr.registerChannel <- worker:
		default:
		}
	} else {
		fmt.Println("FAILURE!!!!!!!!!!!!!!!!!!!!!!!")
		os.Exit(3)
	}
}
