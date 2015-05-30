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

import "strconv"

// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		n, err = fmt.Printf(format, a...)
	}
	return
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
	requestLog map[string]ClientState
	lock       sync.RWMutex
}

func (pb *PBServer) Put(args *PutArgs, reply *PutReply) error {
	// Your code here.
	var err error
	if view.Primary != pb.me {
		reply.Err = ErrWrongServer
		return err
	}
	return pb.ForcePut(args, reply)
}

func (pb *PBServer) ForcePut(args *PutArgs, reply *PutReply) error {
	var err error
	lock.Lock()
	var newVal string
	if args.PutHash {
		curVal := data[args.Key]
		newVal = strconv.Itoa(int(hash(curVal + args.Value)))
	} else {
		newVal = args.Value
	}
	data[args.Key] = newVal
	requestLog[args.Client].lastPut = args.Version
	requestLog
	lock.Unlock()
	pb.forwardRequest(args)
	return nil
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
	// Your code here.
	var err error
	if view.Primary != pb.me {
		reply.Err = ErrWrongServer
		return err
	}
	lock.RLock()
	reply.Value = data[args.Key]
	lock.RUnlock()
	return nil
}

func (pb *PBServer) RecvBackup(args *BackupArgs, reply *BackupReply) error {
	pb.data = args.Data
	pb.requestLog = args.RequestLog
	return nil
}

// ping the viewserver periodically.
func (pb *PBServer) tick() {
	// Your code here.
	args := &PingArgs{pb.me, pb.viewnum}
	var reply PingReply
	ok := call(vs.me, "Clerk.Ping", args, &reply)
	if ok {
		newView := reply.View
		pb.handleBackup(pb.view, newView)
		pb.view = newView
	}
}

func (pb *PBServer) handleBackup(old viewservice.View, newView viewservice.View) {
	if old.Viewnum < newView.Viewnum && old.Primary == pb.me && old.Backup != newView.Backup {
		args := &BackupArgs{data, requestLog}
		var reply BackupReply
		ok := call(newView.Backup, "Clerk.RecvBackup", args, &reply)
		if !ok || reply.Err != OK {
			pb.handleBackup(old, newView)
		}
	}
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
