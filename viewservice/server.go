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
	view        View
	ackedView   View
	idle        string
	liveServers map[string]time.Time
	lastAcks    map[string]uint
	rebooted    map[string]bool
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {

	//DPrintf("Ping from %v with view %d when server's view is %d\n",
	//args.Me, args.Viewnum, vs.view.Viewnum)
	vs.mu.Lock()
	var err error
	err = nil
	vs.updateStats(args)
	vs.ackView(args)
	if args.Me == vs.view.Primary {
		reply.View = vs.view
	} else {
		reply.View = vs.ackedView
	}
	DPrintf("Responded to ping from %v (view %d) with view %s\n", args.Me, args.Viewnum, viewAsString(reply.View))
	vs.mu.Unlock()
	return err
}

// Handles updates related specifically to a received ping:
// - Is this the very first ping that's been received?
// - Has a server rebooted?
func (vs *ViewServer) updateStats(args *PingArgs) {
	if vs.ackedView.Viewnum == 0 {
		DPrintf("Got first ping: primary is %v\n", args.Me)
		vs.view = View{Viewnum: 1, Primary: args.Me}
		vs.ackedView = vs.view
	}
	client := args.Me
	if vs.view.Primary != client && vs.view.Backup != client {
		vs.idle = client
	}
	vs.liveServers[args.Me] = time.Now()
	if args.Viewnum == 0 && vs.lastAcks[args.Me] > 0 {
		DPrintf("Marking %s as rebooted (pinged with %d)...\n", args.Me, args.Viewnum)
		vs.rebooted[args.Me] = true
	} else {
		delete(vs.rebooted, args.Me)
	}
	vs.validateView(vs.isRebooted)
}

// Promote the VS's current view of the world to the blessed (acked) view if the primary has acknowledged
// receipt of it.
func (vs *ViewServer) ackView(args *PingArgs) {
	vs.lastAcks[args.Me] = args.Viewnum

	if vs.lastAcks[vs.view.Primary] == vs.view.Viewnum {
		vs.ackedView = vs.view
		delete(vs.rebooted, vs.view.Primary)
		delete(vs.rebooted, vs.view.Backup)
	}
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {

	// Your code here.
	reply.View = vs.view
	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {

	// Your code here.
	//	DPrintln("Tick...")
	vs.mu.Lock()
	vs.validateView(vs.isDead)
	vs.mu.Unlock()

}

// If a server hasn't pinged in the pre-set amount of time,
// remove it from the list of 'live' servers.
func (vs *ViewServer) markStaleClients() {
	deathCutoff := (DeadPings * PingInterval)
	for client, lastPing := range vs.liveServers {
		if time.Since(lastPing) >= deathCutoff {
			DPrintf("Removing stale server %v\n", client)
			delete(vs.liveServers, client)
		}
	}
}

// convenience type for passing a function to a call to validateView()
type validator func(client string) bool

// Make sure the VS's view accurately represents the state of the world:
// - Sweep out dead clients
// - Promote any idle server to backup if the backup fails its validation check or is absent
// - Promote an initialized backup to primary of the primary fails its validation check
func (vs *ViewServer) validateView(check validator) {
	if vs.view.Viewnum == 0 || vs.ackedView.Viewnum < vs.view.Viewnum {
		return
	}
	vs.markStaleClients()
	//	DPrintf("Before view validation: \n\tCurrent: %s\n\tAcked: %s\n",
	//		viewAsString(vs.view), viewAsString(vs.ackedView))

	backup := vs.view.Backup
	primary := vs.view.Primary
	// backup died, or backup has already been removed but not yet replaced
	if check(backup) || (primary != "" && backup == "" && vs.idle != "") {
		DPrintln("Promoting idle to backup...")
		vs.view = promoteIdle(vs)
	} else {
		// primary died...but we can't promote a newly promoted backup all the way to
		// primary because (insufficiently explained) reasons, so we're only doing this if
		// we haven't just promoted a backup
		if check(primary) {
			DPrintln("Promoting backup to primary...")
			vs.view = promoteBackup(vs)

		}
	}
	//	DPrintf("After view validation: \n\tCurrent: %s\n\tAcked: %s\n",
	//		viewAsString(vs.view), viewAsString(vs.ackedView))
}

// Validation check to see if the client's last ping singalled a crash and reboot
func (vs *ViewServer) isRebooted(client string) bool {
	if client != "" && vs.rebooted[client] {
		delete(vs.rebooted, client)
		return true
	} else {
		return false
	}
}

// Validation check to see if the client failed to sent a timely heartbeat
func (vs *ViewServer) isDead(client string) bool {
	return client != "" && vs.liveServers[client].IsZero()
}

// Promote the backup client to the primary position
func promoteBackup(vs *ViewServer) View {
	idle := vs.idle
	vs.idle = ""
	return View{vs.view.Viewnum + 1, vs.view.Backup, idle}
}

// Promote any idle client to the backup position, leaving the current primary in place
func promoteIdle(vs *ViewServer) View {
	idle := vs.idle
	vs.idle = ""
	return View{vs.view.Viewnum + 1, vs.view.Primary, idle}
}

// Pretty-print a view
func viewAsString(view View) string {
	return fmt.Sprintf("\tNum: %d \tPrimary: %s \tBackup: %s",
		view.Viewnum, view.Primary, view.Backup)
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
	vs.rebooted = make(map[string]bool)

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
