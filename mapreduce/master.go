package mapreduce

import (
	"container/list"
	"fmt"
	//	"math/rand"
)

type WorkerInfo struct {
	address  string
	avaiable bool
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

	// process new registrations
	// go ListenForRegistrations()

	mapChan := make(chan int, mr.nMap)
	mappedChan := make(chan int, mr.nMap)
	redChan := make(chan int, mr.nReduce)
	doneChan := make(chan int)

	for maps := 0; maps < mr.nMap; maps++ {

		// call each available worker at random
		go func(jobNum int) {
			worker := GetNextWorker(mr)
			mapChan <- 1
			repl := &DoJobReply{}
			SendJob(mr, worker, Map, jobNum, len(redChan), repl)
			mappedChan <- 1
		}(maps)
	}

	// reduce jobs
	for reduces := 0; reduces < mr.nReduce; reduces++ {
		go func(jobNum int) {
			<-mappedChan
			worker := GetNextWorker(mr)
			redChan <- 1
			repl := &DoJobReply{}
			SendJob(mr, worker, Reduce, jobNum, len(mappedChan), repl)
			<-redChan
			if jobNum == mr.nReduce {
				doneChan <- 1
			}
		}(reduces)
	}

	<-doneChan
	return mr.KillWorkers()
}

func ListenForRegistrations(mr *MapReduce) {
	//	for mr.alive {
	//		workerAddr := <-mr.registerChannel
	//		mr.Workers[workerAddr] = &WorkerInfo{workerAddr, true}
	//	}
}

func GetNextWorker(mr *MapReduce) string {
	worker := <-mr.registerChannel
	//	fmt.Printf("Found worker: %s", worker)
	// var worker *WorkerInfo
	// for !worker {
	//   if w := mr.Workers[rand.Intn(len(mr.Workers))]; w.available {
	//     worker = w
	//   }
	// }
	return worker
}

func SendJob(mr *MapReduce, worker string, op JobType, jobNum int, others int, repl *DoJobReply) {
	args := &DoJobArgs{
		File:          mr.file,
		Operation:     op,
		JobNumber:     jobNum,
		NumOtherPhase: others,
	}
	fmt.Printf("Sending job %v %d to %s...\n", op, jobNum, worker)
	call(worker, "Worker.DoJob", args, repl)
	mr.registerChannel <- worker
}
