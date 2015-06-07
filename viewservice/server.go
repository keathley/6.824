package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "fmt"
import "os"

type ViewServer struct {
	mu   sync.Mutex
	l    net.Listener
	dead bool
	me   string

	currentView View
	nextView    View
	acked       bool
	primary     time.Time
	backup      time.Time
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if (vs.currentView.Viewnum == 0) {
		// first server is automatically the primary
		vs.currentView = View{
			Viewnum:1,
			Primary:args.Me,
			Backup: "",
		}

		// initialize the nextView
		vs.nextView = vs.currentView
		vs.primary = time.Now()

	} else if (args.Me == vs.currentView.Primary) {
		// is the primary acking the current view?
		if (args.Viewnum == vs.currentView.Viewnum) {
			vs.primary = time.Now()
			vs.acked = vs.currentView.Viewnum == vs.nextView.Viewnum
			// mismatch means advance the view
			if (!vs.acked) {
				vs.currentView = vs.nextView
			}
		} else {
			// primary's out of sync; promote the backup
			vs.promote()
		}

	} else if (args.Me == vs.currentView.Backup && args.Viewnum == vs.currentView.Viewnum) {
		// the backup is checking in; update checkin time
		// accordingly: if he's being promoted, update primary
		// time. otherwise, keep the backup time up to date
		if (args.Me == vs.nextView.Primary && vs.acked) {
			vs.currentView = vs.nextView
			vs.primary = time.Now()
			vs.acked = false
		} else {
			vs.backup = time.Now()
		}

	} else if (args.Me != vs.nextView.Primary && vs.nextView.Backup == "") {
		// if there's no backup, promote the first
		// non-allocated server to be the backup
		vs.nextView.Viewnum = vs.currentView.Viewnum + 1
		vs.nextView.Backup = args.Me
		vs.backup = time.Now()
	}

	reply.View = vs.currentView
	return nil
}

func (vs *ViewServer) promote() {
	// no lock necessary as this is always called in a safe context
	vs.nextView = View{
		Viewnum:vs.currentView.Viewnum+1,
		Primary:vs.currentView.Backup,
		Backup:"",}

	// last check-in by the new primary
	vs.primary = vs.backup
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	reply.View = vs.currentView
	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	since := time.Since(vs.primary)
	if (since > MaxPingTime) {
		vs.promote()
	}
}

//
// tell the server to shut itself down.
// for testing.
// please don't change this function.
//
func (vs *ViewServer) Kill() {
	vs.dead = true
	vs.l.Close()
}

func StartServer(me string) *ViewServer {
	vs := new(ViewServer)
	vs.me = me
	// Your vs.* initializations here.

	// tell net/rpc about our RPC server and handlers.
	rpcs := rpc.NewServer()
	rpcs.Register(vs)

	// prepare to receive connections from clients.
	// change "unix" to "tcp" to use over a network.
	os.Remove(vs.me) // only needed for "unix"
	l, e := net.Listen("unix", vs.me)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	vs.l = l

	// please don't change any of the following code,
	// or do anything to subvert it.

	// create a thread to accept RPC connections from clients.
	go func() {
		for vs.dead == false {
			conn, err := vs.l.Accept()
			if err == nil && vs.dead == false {
				go rpcs.ServeConn(conn)
			} else if err == nil {
				conn.Close()
			}
			if err != nil && vs.dead == false {
				fmt.Printf("ViewServer(%v) accept: %v\n", me, err.Error())
				vs.Kill()
			}
		}
	}()

	// create a thread to call tick() periodically.
	go func() {
		for vs.dead == false {
			vs.tick()
			time.Sleep(PingInterval)
		}
	}()

	return vs
}
