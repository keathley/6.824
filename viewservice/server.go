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

	// Your declarations here.
	View View;
	NextView View;
	primaryMisses uint;
	backupMisses uint;
	acked bool;
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {
	// first server to check in becomes primary
	if (vs.View.Viewnum == 0) {
		vs.View = View{Viewnum:1, Primary:args.Me, Backup:""}
		vs.NextView = vs.View

	// ping from primary resets timeout and acks view
	} else if (args.Me == vs.View.Primary) {

		// if primary is up-to-date, reset clock and ack view
		if (args.Viewnum == vs.View.Viewnum) {
			vs.primaryMisses = 0
			if (vs.View.Viewnum < vs.NextView.Viewnum) {
				vs.View = vs.NextView
				vs.acked = false
			} else {
				vs.acked = true
			}

		} else if (args.Viewnum == 0) {
			primaryFailure(vs)
		}

	// ping from backup resets backup timeout
	} else if (args.Me == vs.View.Backup && args.Viewnum == vs.View.Viewnum) {
		vs.backupMisses = 0

		if (args.Me == vs.NextView.Primary && vs.acked) {
			vs.View = vs.NextView
			vs.primaryMisses = 0
			vs.acked = false
		}

	// if we don't have a backup and a new server comes in, store it as backup
	} else if (args.Me != vs.NextView.Primary && vs.NextView.Backup == "") {
		vs.NextView = View{Viewnum:vs.View.Viewnum+1, Primary:vs.NextView.Primary, Backup:args.Me}
		vs.backupMisses = 0
	}

	reply.View = vs.View

	return nil
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {

	reply.View = vs.View

	return nil
}

func primaryFailure(vs *ViewServer) {
	vs.NextView = View{Viewnum:vs.View.Viewnum+1, Primary:vs.View.Backup, Backup:""}
	vs.primaryMisses = 0
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	vs.primaryMisses += 1
	vs.backupMisses += 1

	if (vs.primaryMisses > DeadPings) {
		primaryFailure(vs)
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
