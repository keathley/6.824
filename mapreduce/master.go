package mapreduce

import "container/list"

type WorkerInfo struct {
	address string
	// You can add definitions here.
}

// Clean up all workers by sending a Shutdown RPC to each one of them Collect
// the number of jobs each work has performed.
func (mr *MapReduce) KillWorkers() *list.List {
	l := list.New()
	for _, w := range mr.Workers {
		DPrintf(3, "DoWork: shutdown %s\n", w.address)
		args := &ShutdownArgs{}
		var reply ShutdownReply
		ok := call(w.address, "Worker.Shutdown", args, &reply)
		if ok == false {
			DPrintf(1, "DoWork: RPC %s shutdown error\n", w.address)
		} else {
			l.PushBack(reply.Njobs)
		}
	}
	return l
}

func (mr *MapReduce) RunMaster() *list.List {
	go func() {
		for worker := range mr.registerChannel {
			go func(worker string) {
				DPrintf(1, "Launch Worker: %s\n", worker)
				for {
					job := <-mr.jobChannel
					done := call(worker, "Worker.DoJob", job, &DoJobReply{})
					if !done {
						mr.jobChannel <- job
						DPrintf(1, "Worker Failed: %s | %s #%d\n", worker, job.Operation, job.JobNumber)
						return
					}
				}
			}(worker)
		}
	}()
	mr.runJobs(Map, mr.nMap, mr.nReduce)
	mr.runJobs(Reduce, mr.nReduce, mr.nMap)
	return mr.KillWorkers()
}

func (mr *MapReduce) runJobs(phase JobType, nJobs int, nOtherJobs int) {
	DPrintf(1, "Start %s phase (%d jobs)\n", phase, nJobs)
	for job := 0; job < nJobs; job++ {
		mr.jobChannel <- &DoJobArgs{
			File:          mr.file,
			Operation:     phase,
			NumOtherPhase: nOtherJobs,
			JobNumber:     job,
		}
		DPrintf(2, "Put Map %d onto JobChannel\n", job)
	}
	DPrintf(1, "%s phase complete\n", phase)
}
