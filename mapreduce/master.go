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

func (mr *MapReduce) RunMaster() *list.List {
	for mapJobNum := 0; mapJobNum < mr.nMap; mapJobNum++ {
		var jobArgs = DoJobArgs{
			File: mr.file,
			Operation: Map,
			NumOtherPhase: mr.nReduce,
			JobNumber: mapJobNum,
		}
		go mr.QueueJob(&jobArgs)
	}

	// wait for all map jobs to be done before moving on
	mr.awaitJobs.Wait()

	mr.awaitJobs.Add(mr.nReduce)
	for reduceJobNum := 0; reduceJobNum < mr.nReduce; reduceJobNum++ {
		var reduceJobArgs = DoJobArgs{
			File: mr.file,
			Operation: Reduce,
			NumOtherPhase: mr.nMap,
			JobNumber: reduceJobNum,
		}
		go mr.QueueJob(&reduceJobArgs)
	}

	fmt.Println("All reducer jobs queued")

	// wait for all reducers to complete before killing workers
	mr.awaitJobs.Wait()
	fmt.Println("All reducer jobs done")

	return mr.KillWorkers()
}

func (mr *MapReduce) QueueJob(args *DoJobArgs) {
	var reply = DoJobReply{}

	// tell a worker to execute job
	worker := <-mr.registerChannel
	done := call(worker, "Worker.DoJob", &args, &reply)

	// there are two different types of failure to investigate:
	// the RPC call could have failed (see done's state) or the
	// worker itself could have failed (see reply.OK)
	if done && reply.OK {
		mr.awaitJobs.Done()

		// requeue worker as available
		mr.registerChannel <- worker
	} else {
		// log failure types separately for debugging
		fmt.Printf("MINE:RPC failure: %t. Worker failure: %t\n", done, reply.OK)
		go mr.QueueJob(args)
	}
}
