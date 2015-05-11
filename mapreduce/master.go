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

	fmt.Println("All mapping jobs queued")

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
	mr.awaitReducers.Wait()
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
		if args.Operation == Reduce {
			// this reducer has completed, so we can mark
			// it as done
			mr.awaitReducers.Done()
		}

		// requeue worker as available
		mr.registerChannel <- worker
	} else {
		// log failure types separately for debugging
		fmt.Printf("RPC failure: %t. Worker failure: %t\n", done, reply.OK)
		// potential problem: since we're not restarting a job
		// when a worker fails, we have a potential to wait
		// forever (because the reducers might not all check
		// in). hopefully, part 3 addresses this concern
	}
}
