package pbservice

import "net"
import "fmt"
import "net/rpc"
import "log"
import "time"
import "github.com/keathley/6.824/viewservice"
import "os"
import "syscall"
import "math/rand"
import "sync"
import "errors"

import "strconv"

// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		n, err = fmt.Printf(format, a...)
	}
	return
}

// Pretty-print a view
func viewAsString(view viewservice.View) string {
	return fmt.Sprintf("\tNum: %d \tPrimary: %s \tBackup: %s",
		view.Viewnum, view.Primary, view.Backup)
}

func printMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("%s -> %s\n", k, v)
	}
}

type PBServer struct {
	l          net.Listener
	dead       bool // for testing
	unreliable bool // for testing
	me         string
	vs         *viewservice.Clerk
	done       sync.WaitGroup
	finish     chan interface{}
	view       viewservice.View // first custom field
	data       map[string]string
	requestLog map[string]*ClientState
	lock       sync.RWMutex
}

func (pb *PBServer) isPrimary() bool {
	return pb.getView().Primary == pb.me
}

func (pb *PBServer) isBackup() bool {
	return pb.getView().Backup == pb.me
}

func (pb *PBServer) Put(args *PutArgs, reply *PutReply) error {
	var err error
	if !pb.isPrimary() {
		DPrintf("WARN: Attempted to put %s -> %s to non-primary (%s)\n", args.Key, args.Value, pb.me)
		reply.Err = ErrWrongServer
		return err
	}
	dupe, dupeErr := pb.filterDuplicate(args, reply)
	if dupeErr == ErrWrongServer {
		reply.Err = dupeErr
	} else if !dupe {
		err = pb.DoPut(args, reply)
	}
	//	DPrintf("Reply in Put: %s\n", reply)
	return err
}

func (pb *PBServer) BackupPut(args *PutArgs, reply *PutReply) error {
	var err error
	if pb.isPrimary() {
		reply.Err = ErrWrongServer
		return err
	}
	dupe, dupeErr := pb.filterDuplicate(args, reply)
	if dupeErr == ErrWrongServer {
		reply.Err = dupeErr
	} else if !dupe {
		err = pb.DoPut(args, reply)
	}
	return err
}

func (pb *PBServer) DoPut(args *PutArgs, reply *PutReply) error {
	DPrintf("%s Putting %s -> %s for %s to server %s (version %d; hash? %v)\n", pb.me, args.Key, args.Value, args.Client, pb.me, args.Version, args.DoHash)
	var err error
	var newVal string
	curVal, _ := pb.getVal(args.Key)
	if args.DoHash {
		newVal = strconv.Itoa(int(hash(curVal + args.Value)))
		DPrintf("Hashed value for '%s' + '%s': '%s'\n", curVal, args.Value, newVal)
	} else {
		newVal = args.Value
	}
	var fReply = &PutReply{}
	if !pb.forwardRequest(args, fReply) {
		err = errors.New("Could not forward request to backup")
	} else if fReply.PreviousValue == NotBackup || fReply.PreviousValue == curVal {
		reply.PreviousValue = pb.setVal(args.Key, newVal)
		reply.Err = OK
		pb.updateCache(args, reply)
	}
	//	DPrintf("Reply in DoPut: %s\n", reply)
	return err
}

func (pb *PBServer) getVal(key string) (string, bool) {
	pb.lock.RLock()
	val, ok := pb.data[key]
	pb.lock.RUnlock()
	return val, ok
}

func (pb *PBServer) setVal(key string, val string) string {
	pb.lock.Lock()
	prev := pb.data[key]
	pb.data[key] = val
	pb.lock.Unlock()
	return prev
}

func (pb *PBServer) forwardRequest(args *PutArgs, reply *PutReply) bool {
	var ok = true
	reply.PreviousValue = NotBackup
	for !(pb.isBackup() || pb.getView().Backup == "") {
		ok = call(pb.getView().Backup, "PBServer.BackupPut", args, reply)
		if reply.Err == OK {
			break
		}
	}
	return ok
}

func (pb *PBServer) forwardRequest(args *PutArgs) {
	if view.Backup != "" {
		var reply PutReply
		ok := call(view.Backup, "PBServer.ForcePut", args, &reply)
		if !ok || reply.Err != OK {
			pb.forwardRequest(args)
		}
	}
}

func (pb *PBServer) Get(args *GetArgs, reply *GetReply) error {
	DPrintf("%s Getting %s for %s from server %s (version %d)\n", pb.me, args.Key, args.Client, pb.me, args.Version)
	//	DPrintf("Data:\n")
	//	printMap(pb.data)
	var err error
	if !pb.isPrimary() {
		DPrintf("WARN: Attempted to get %s from non-primary (%s)\n", args.Key, pb.me)
		reply.Err = ErrWrongServer
	} else {
		dupe, dupeErr := pb.filterDuplicate(args, reply)
		if dupeErr == ErrWrongServer {
			reply.Err = dupeErr
		} else {
			if !dupe {
				val, ok := pb.getVal(args.Key)
				reply.Value = val // ok doesn't matter here; the zero-value for string is fine
				if !ok {
					reply.Err = ErrNoKey
				} else {
					reply.Err = OK
				}
				pb.updateCache(args, reply)
			}
		}
	}
	DPrintf("Got %s\n", reply.Value)
	return err
}

func (pb *PBServer) updateCache(args interface{}, reply interface{}) {
	//	DPrintf("Updating request cache...\n")
	var lastState *ClientState
	var cacheField *State
	var version int
	var client, val string
	input, ok := args.(*GetArgs)
	output, ok := reply.(*GetReply)
	if ok {
		client = input.Client
		cacheField = lastState.Get
		version = input.Version
		val = output.Value
	} else {
		input := args.(*PutArgs)
		client = input.Client
		cacheField = lastState.Put
		output := reply.(*PutReply)
		version = input.Version
		val = output.PreviousValue
	}
	pb.putCache(client, &cacheField, version, val)
	//	DPrintf("New cache state: %s\n", lastState)
}

func (pb *PBServer) putCache(client string, field **State, version int, val string) {
	state := pb.getOrCreateCacheState(client)
	pb.lock.Lock()
	state.field.Version = version
	state.field.Value = val
	pb.requestLog[client] = state
	pb.lock.Unlock()
}

// Filter duplicate requests, filling in the reply with the cached value if the request is determined to be a duplicate. Returns a bool and a string:
// The bool represents whether the request was a duplicate; the string will be populated with an error message if one occurred.
func (pb *PBServer) filterDuplicate(args interface{}, reply interface{}) (bool, Err) {
	input, ok := args.(*GetArgs)
	//	DPrintf("Filtering for duplicate request...\n")
	var duplicate bool
	var cacheField *State
	var version int
	var lastState *ClientState
	if ok {
		lastState = pb.getOrCreateCacheState(input.Client)
		cacheField = lastState.Get
		version = input.Version
	} else {
		input := args.(*PutArgs)
		lastState = pb.getOrCreateCacheState(input.Client)
		cacheField = lastState.Put
		version = input.Version
	}
	//	DPrintf("Incoming request version: %d; cached request version: %d\n", version, cacheField.Version)
	if cacheField.Version-version > 1 {
		// version gap, indicative of a network partition
		DPrintf("Version gap found: %d versions apart\n", cacheField.Version-version)
		return false, ErrWrongServer
	}
	duplicate = (cacheField.Version == version)
	if duplicate {
		output, ok := reply.(*GetReply)
		// it doesn't seem like I should have to repeat myself here, but the Go compiler won't let me
		// redeclare 'output' inside a block and use the new value outside
		if !ok {
			output := reply.(*PutReply)
			output.PreviousValue = cacheField.Value
			output.Err = OK
			DPrintf("Duplicate put detected at %s; returning '%s'\n", pb.me, output.PreviousValue)
		} else {
			output.Value = cacheField.Value
			output.Err = OK
			DPrintf("Duplicate get detected at %s; returning '%s'\n", pb.me, output.Value)
		}

	}
	return duplicate, OK
}

func (pb *PBServer) getOrCreateCacheState(client string) *ClientState {
	pb.lock.RLock()
	state := pb.requestLog[client]
	pb.lock.RUnlock()
	if state == nil {
		state = &ClientState{}
		state.Get = &State{}
		state.Put = &State{}
	}
	//	DPrintf("State found in cache: %s\n", state)
	return state
}

// Receive a backup of the current data from the primary server
func (pb *PBServer) RecvBackup(args *BackupArgs, reply *BackupReply) error {
	DPrintf("Initializing backup data...\n")
	pb.establishView()
	if pb.isPrimary() {
		DPrintf("Backup request sent to server that thinks it's the primary\n")
		reply.Err = ErrWrongServer
		return nil
	}
	pb.lock.Lock()
	pb.data = args.Data
	pb.requestLog = args.RequestLog
	pb.lock.Unlock()
	reply.Err = OK
	return nil
}

// ping the viewserver periodically.
func (pb *PBServer) tick() {
	pb.establishView()
}

func (pb *PBServer) establishView() {
	curView := pb.getView()
	view, err := pb.vs.Ping(curView.Viewnum)
	if err == nil {
		//		newView := reply.View
		if !pb.handleBackup(curView, view) {
			pb.establishView()
			return
		}
		pb.setView(view)
	} else {
		// if we can't talk to the viewserver, we're partitioned and can't be trusted
		DPrintf("Server %s partitioned; resetting view\n", pb.me)
		pb.setView(viewservice.View{})
	}
	DPrintf("View for %s: %s\n", pb.me, viewAsString(curView))
}

func (pb *PBServer) getView() viewservice.View {
	pb.lock.RLock()
	view := pb.view
	pb.lock.RUnlock()
	return view
}

func (pb *PBServer) setView(view viewservice.View) {
	pb.lock.Lock()
	pb.view = view
	pb.lock.Unlock()
}

func (pb *PBServer) handleBackup(old viewservice.View, newView viewservice.View) bool {
	if old.Viewnum < newView.Viewnum &&
		old.Primary == pb.me && old.Backup != newView.Backup {
		var reply BackupReply
		pb.lock.RLock()
		args := &BackupArgs{pb.data, pb.requestLog}
		ok := call(newView.Backup, "PBServer.RecvBackup", args, &reply)
		ok = ok && reply.Err == OK
		pb.lock.RUnlock()
		return ok
	}
	return true
}

// tell the server to shut itself down.
// please do not change this function.
func (pb *PBServer) kill() {
	pb.dead = true
	pb.l.Close()
}

func StartServer(vshost string, me string) *PBServer {
	pb := new(PBServer)
	pb.me = me
	pb.vs = viewservice.MakeClerk(me, vshost)
	pb.finish = make(chan interface{})
	// Your pb.* initializations here.

	pb.data = make(map[string]string)
	pb.requestLog = make(map[string]*ClientState)

	rpcs := rpc.NewServer()
	rpcs.Register(pb)

	os.Remove(pb.me)
	l, e := net.Listen("unix", pb.me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	pb.l = l

	// please do not change any of the following code,
	// or do anything to subvert it.

	go func() {
		for pb.dead == false {
			conn, err := pb.l.Accept()
			if err == nil && pb.dead == false {
				if pb.unreliable && (rand.Int63()%1000) < 100 {
					// discard the request.
					conn.Close()
				} else if pb.unreliable && (rand.Int63()%1000) < 200 {
					// process the request but force discard of reply.
					c1 := conn.(*net.UnixConn)
					f, _ := c1.File()
					err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
					if err != nil {
						fmt.Printf("shutdown: %v\n", err)
					}
					pb.done.Add(1)
					go func() {
						rpcs.ServeConn(conn)
						pb.done.Done()
					}()
				} else {
					pb.done.Add(1)
					go func() {
						rpcs.ServeConn(conn)
						pb.done.Done()
					}()
				}
			} else if err == nil {
				conn.Close()
			}
			if err != nil && pb.dead == false {
				fmt.Printf("PBServer(%v) accept: %v\n", me, err.Error())
				pb.kill()
			}
		}
		DPrintf("%s: wait until all request are done\n", pb.me)
		pb.done.Wait()
		// If you have an additional thread in your solution, you could
		// have it read to the finish channel to hear when to terminate.
		close(pb.finish)
	}()

	pb.done.Add(1)
	go func() {
		for pb.dead == false {
			pb.tick()
			time.Sleep(viewservice.PingInterval)
		}
		pb.done.Done()
	}()

	return pb
}
