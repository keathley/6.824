package mapreduce

import "container/list"
import "fmt"

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

func RPCTask(mr *MapReduce, worker string, jobNumber int, phaseNumber int, operation JobType) bool {
	var args DoJobArgs
	var reply DoJobReply
	args.NumOtherPhase = phaseNumber
	args.Operation = operation
	args.File = mr.file
	args.JobNumber = jobNumber
	return call(worker, "Worker.DoJob", args, &reply)
}

func mapJob(mr *MapReduce, jobNumber int) (bool, string) {
	var worker string
	var success bool
	select {
	case worker = <-mr.freeChannel:
		success = RPCTask(mr, worker, jobNumber, mr.nReduce, Map)
	case worker = <-mr.registerChannel:
		success = RPCTask(mr, worker, jobNumber, mr.nReduce, Map)
	}
	return success, worker
}

func reduceJob(mr *MapReduce, jobNumber int) (bool, string) {
	var worker string
	var success bool

	select {
	case worker = <-mr.freeChannel:
		success = RPCTask(mr, worker, jobNumber, mr.nMap, Reduce)
	case worker = <-mr.registerChannel:
		success = RPCTask(mr, worker, jobNumber, mr.nMap, Reduce)
	}
	return success, worker
}

func readFromChannel(c chan int) {
	<-c
	return
}

func (mr *MapReduce) RunMaster() *list.List {
	// Your code here

	var mapChannel = make(chan int, mr.nMap)
	var reduceChannel = make(chan int, mr.nReduce)

	for num := 0; num < mr.nMap; num++ {
		go func(jobNumber int) {
			for {
				var OK, worker = mapJob(mr, jobNumber)
				if OK {
					mapChannel <- jobNumber
					mr.freeChannel <- worker
					return
				}
			}
		}(num)
	}

	// failing on the last file doing it this way
	// defer close(mapChannel)

	for mapJobNumber := 0; mapJobNumber < mr.nMap; mapJobNumber++ {
		readFromChannel(mapChannel)
	}

	for num := 0; num < mr.nReduce; num++ {
		go func(jobNumber int) {
			for {
				var OK, worker = reduceJob(mr, jobNumber)
				if OK {
					reduceChannel <- jobNumber
					mr.freeChannel <- worker
					return
				}
			}
		}(num)
	}

	for reduceJobNumber := 0; reduceJobNumber < mr.nReduce; reduceJobNumber++ {
		readFromChannel(reduceChannel)
	}

	return mr.KillWorkers()
}
