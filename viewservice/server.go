package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "fmt"
import "os"

//import "errors"

type ViewServer struct {
	mu   sync.Mutex
	l    net.Listener
	dead bool
	me   string

	// Your declarations here.
	view        View
	ackedView   View
	idle        string
	liveServers map[string]time.Time
	lastAcks    map[string]uint
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {

	// Your code here.
	var err error
	err = nil
	fmt.Printf("Ping from %v with view %d when server's view is %d\n", args.Me, args.Viewnum, vs.view.Viewnum)
	if vs.ackedView.Viewnum == 0 {
		fmt.Printf("Got first ping: primary is %v\n", args.Me)
		vs.view = View{Viewnum: 1, Primary: args.Me}
		vs.ackedView = vs.view
	}

	reply.View = vs.ackedView
	vs.updateStats(args)
	client := args.Me
	if vs.view.Primary != client && vs.view.Backup != client {
		vs.idle = client
	}
	// err := errors.New("blah")
	return err
}

func (vs *ViewServer) updateStats(args *PingArgs) {
	vs.liveServers[args.Me] = time.Now()
	vs.lastAcks[args.Me] = args.Viewnum
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {

	// Your code here.
	reply.View = vs.ackedView
	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	// The view service proceeds to a new view when either it hasn't received a Ping from the primary or backup for DeadPingsPingIntervals,
	//or if the primary or backup crashed and restarted,
	// or if there is no backup and there's an idle server (a server that's been Pinging but is neither the primary nor the backup).
	//But the view service must not change views (i.e., return a different view to callers) until the primary from the current view acknowledges that it is operating in the current view (by sending a Ping with the current view number). If the view service has not yet received an acknowledgment for the current view from the primary of the current view, the view service should not change views even if it thinks that the primary or backup has died.

	// Your code here.
	vs.mu.Lock()
	vs.removeStaleClients()
	vs.validateView()
	vs.mu.Unlock()
}

func (vs *ViewServer) removeStaleClients() {
	deathCutoff := (DeadPings * PingInterval)
	for client, lastPing := range vs.liveServers {
		if time.Since(lastPing) >= deathCutoff {
			fmt.Printf("Removing stale server %v\n", client)
			delete(vs.liveServers, client)
		}
	}
}

func (vs *ViewServer) validateView() {
	if vs.view.Viewnum == 0 {
		return
	}

	var advanced = false
	// backup rebooted
	if vs.view.Backup != "" && vs.rebooted(vs.view.Backup) {
		vs.view.Backup = vs.idle
		vs.idle = ""
		advanced = true
	}
	// primary rebooted
	if vs.view.Primary != "" && vs.rebooted(vs.view.Primary) {

		vs.view.Primary = vs.view.Backup
		if vs.idle != "" {
			vs.view.Backup = vs.idle
			vs.idle = ""
		} else {
			vs.view.Backup = ""
		}
		advanced = true

	}
	// filling an empty backup slot
	if vs.view.Primary != "" && vs.view.Backup == "" && vs.idle != "" {
		vs.view.Backup = vs.idle
		vs.idle = ""
		advanced = true
	}
	var acked = vs.lastAcks[vs.view.Primary] == vs.view.Viewnum
	if advanced {
		vs.view.Viewnum += 1
	}
	if acked {
		vs.ackedView = vs.view
	}
	fmt.Printf("After view validation: \tCurrent: %d\tAcked: %d\n", vs.view.Viewnum, vs.ackedView.Viewnum)
}

func (vs *ViewServer) rebooted(client string) bool {
	//	fmt.Printf("Liveness check for %s: lastPing at %v; last acked view: %d\n", client[len(client)-6:len(client)], vs.liveServers[client], vs.lastAcks[client])
	return vs.liveServers[client].IsZero() || (vs.lastAcks[client] == 0 && vs.lastAcks[client] < vs.view.Viewnum)
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
	vs.view = View{Viewnum: 0}
	vs.ackedView = View{Viewnum: 0}
	vs.liveServers = make(map[string]time.Time)
	vs.lastAcks = make(map[string]uint)

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
