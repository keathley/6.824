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
	go mr.flyMyPretties()
	mr.giterdone(Map, mr.nMap)
	mr.giterdone(Reduce, mr.nReduce)
	return mr.KillWorkers()
}

func (mr *MapReduce) flyMyPretties() {
	for worker := range mr.registerChannel {
		DPrintf(1, "Launch Worker: %s\n", worker)
		go mr.werkit(worker)
	}
}

func (mr *MapReduce) werkit(worker string) {
	for fails := 0; fails < FAILURES; {
		job := <-mr.jobChannel
		done := call(worker, "Worker.DoJob", job, &DoJobReply{})
		if done {
			mr.jobDoneChannel <- job
		} else {
			DPrintf(1, "Fail: %s | %s #%d\n", worker, job.Operation, job.JobNumber)
			fails++
		}
	}
}

func (mr *MapReduce) giterdone(phase JobType, nJobs int) {
	DPrintf(1, "Start %s phase (%d jobs)\n", phase, nJobs)
	jobsDone := make([]bool, nJobs)
	for i := 0; i < nJobs*ATTEMPTS; {
		job := i % nJobs
		if jobsDone[job] {
			i++
			continue
		}
		select {
		case mr.jobChannel <- mr.jobArgs(phase, job):
			DPrintf(2, "%d:Put Map %d onto JobChannel\n", i, job)
			i++
		case donejob := <-mr.jobDoneChannel:
			if donejob.Operation == phase {
				DPrintf(2, "Job Done: %s #%d\n", phase, donejob.JobNumber)
				jobsDone[donejob.JobNumber] = true
			}
		}
	}
	DPrintf(1, "%s phase complete\n", phase)
}

func (mr *MapReduce) jobArgs(phase JobType, job int) *DoJobArgs {
	return &DoJobArgs{
		File:          mr.file,
		Operation:     phase,
		NumOtherPhase: mr.otherPhaseJobs(phase),
		JobNumber:     job,
	}
}

func (mr *MapReduce) otherPhaseJobs(phase JobType) int {
	if phase == Map {
		return mr.nReduce
	}
	return mr.nMap
}
