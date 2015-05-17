package mapreduce

import "container/list"

type WorkerInfo struct {
	address string
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
				if job := <-mr.jobs; call(worker, "Worker.DoJob", job, &DoJobReply{}) {
					mr.done <- job
					mr.registerChannel <- worker
				} else {
					mr.jobs <- job
				}
			}(worker)
		}
	}()
	mr.jobFactory(Map, mr.nMap, mr.nReduce)
	mr.jobFactory(Reduce, mr.nReduce, mr.nMap)
	return mr.KillWorkers()
}

func (mr *MapReduce) jobFactory(phase JobType, nJobs int, nOtherJobs int) {
	for job := 0; job < nJobs; job++ {
		mr.jobs <- &DoJobArgs{
			File:          mr.file,
			Operation:     phase,
			NumOtherPhase: nOtherJobs,
			JobNumber:     job,
		}
		defer func() { <-mr.done }()
	}
}
