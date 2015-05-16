package mapreduce

import (
	"container/list"
	"fmt"
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
		mr.mapJobChan <- mapJobId
	}

	for reduceJobId := 0; reduceJobId < mr.nReduce; reduceJobId++ {
		mr.reduceJobChan <- reduceJobId
	}

	for mapJobId := range mr.mapJobChan {
		go mr.SendMapJob(mapJobId)
	}

	for reduceJobId := range mr.reduceJobChan {
		go mr.SendReduceJob(reduceJobId)
	}

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

	if ok == false {
		mr.handleWorkerFailure(operation, jobId)
	} else {
		go mr.putWorkerBackInQueue(worker)
		mr.addToJobCompleteChannel(operation, jobId)
		if mr.workingChannelIsDone(operation) {
			mr.closeWorkingChannel(operation)
		}
	}
}

func (mr *MapReduce) putWorkerBackInQueue(worker string) {
	args := &RegisterArgs{Worker: worker}
	reply := &RegisterReply{}
	call(mr.MasterAddress, "MapReduce.Register", args, reply)
}

func (mr *MapReduce) addToJobCompleteChannel(operation JobType, jobId int) {
	switch operation {
	case Map:
		mr.mapJobCompleteChan <- jobId
	case Reduce:
		mr.reduceJobCompleteChan <- jobId
	}
}

func (mr *MapReduce) workingChannelIsDone(operation JobType) bool {
	switch operation {
	case Map:
		return len(mr.mapJobCompleteChan) >= mr.nMap
	case Reduce:
		return len(mr.reduceJobCompleteChan) >= mr.nReduce
	}

	return false
}

func (mr *MapReduce) handleWorkerFailure(operation JobType, jobId int) {
	switch operation {
	case Map:
		mr.mapJobChan <- jobId
	case Reduce:
		mr.reduceJobChan <- jobId
	}
}

func (mr *MapReduce) closeWorkingChannel(operation JobType) {
	switch operation {
	case Map:
		close(mr.mapJobChan)
	case Reduce:
		close(mr.reduceJobChan)
	}
}
