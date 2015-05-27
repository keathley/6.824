package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "fmt"
import "os"

type ViewServer struct {
	mu       sync.Mutex
	l        net.Listener
	dead     bool
	me       string
	view     View
	lastPing map[string]time.Time
	ack      uint
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {
	vs.lastPing[args.Me] = time.Now()
	if args.Me == vs.view.Primary && args.Viewnum == 0 {
		vs.death(args.Me)
	} else if args.Me == vs.view.Primary {
		vs.ack = args.Viewnum
	} else if vs.view.Primary == "" {
		vs.view.Primary = args.Me
		vs.view.Viewnum += 1
	} else if vs.view.Backup == "" {
		vs.view.Backup = args.Me
		vs.view.Viewnum += 1
	}
	reply.View = vs.view
	return nil
}

func (vs *ViewServer) death(server string) {
	if vs.ack == vs.view.Viewnum {
		delete(vs.lastPing, server)
		if server == vs.view.Primary || server == vs.view.Backup {
			if server == vs.view.Primary {
				vs.view.Primary = vs.view.Backup
			}
			vs.view.Backup = ""
			for s := range vs.lastPing {
				if s != vs.view.Primary {
					vs.view.Backup = s
				}
			}
			vs.view.Viewnum += 1
		}
	}
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {
	reply.View = vs.view
	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	for server, pingTime := range vs.lastPing {
		if time.Now().Sub(pingTime) > DeadPings*PingInterval {
			vs.death(server)
		}
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
	vs.lastPing = make(map[string]time.Time)

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
