package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "fmt"
import "os"

const (
	Idle    = "Idle"
	Primary = "Primary"
	Backup  = "Backup"
)

type ServerStatus string

type Server struct {
	Me       string
	PingTime time.Time
	Status   ServerStatus
	ack      bool
}

func (s *Server) PromoteTo(status ServerStatus) {
	s.Status = status
}

func (s *Server) Ack() {
	s.ack = true
}

func (s *Server) HasAcked() (ack bool) {
	return
}

type AvailableServers struct {
	servers []Server
}

func (as *AvailableServers) Add(name string) bool {
	_, isFound := as.FindServer(name)

	if !isFound {
		server := Server{
			Me:       name,
			Status:   Idle,
			PingTime: time.Now(),
		}
		as.servers = append(as.servers, server)
	}
	return false
}

//
// Updates the server.  If it's a new server then it adds it to the list
// This traverses the list in linear time so don't add many servers :)
//
func (as *AvailableServers) Update(name string) bool {
	server, found := as.FindServer(name)
	if found {
		server.PingTime = time.Now()
	}

	return false
}

//
// Finds the first server that matches the name
//
func (as *AvailableServers) FindServer(name string) (server *Server, found bool) {
	for i, s := range as.servers {
		if s.Me == name {
			found = true
			server = &as.servers[i]
		}
	}

	return
}

//
// Returns the length of the servers list
//
func (as *AvailableServers) Len() int {
	return len(as.servers)
}

//
// Returns the first Idle server in the list
//
func (as *AvailableServers) GetIdle() (server *Server, found bool) {
	for i, s := range as.servers {
		if s.Status == Idle {
			found = true
			server = &as.servers[i]
		}
	}

	return
}

//
// Removes servers from the list that have larger ping times then the
// total allowed dead time
//
func (as *AvailableServers) KillDeadServers() (deadServers []Server, didKillServer bool) {
	newServers := as.servers[:0]
	deadServers = make([]Server, 0, cap(as.servers))

	for _, server := range as.servers {
		elapsedTime := time.Since(server.PingTime)
		if elapsedTime <= TotalDeadTime {
			newServers = append(newServers, server)
		} else {
			deadServers = append(deadServers, server)
		}
	}
	as.servers = newServers
	didKillServer = len(deadServers) > 0

	// fmt.Println("Servers", as.servers)
	// fmt.Println("Dead Servers", deadServers)
	// fmt.Println("Did Servers Die", didKillServer)

	return
}

//
// Adds the server as a primary to the View and increments the Viewnum
//
func (v *View) PromotePrimary(server Server) {
	v.Primary = server.Me
	v.Viewnum += 1
}

//
// Add the server as the backup and increments the Viewnum
//
func (v *View) PromoteBackup(server Server) {
	v.Backup = server.Me
	v.Viewnum += 1
}

//
// demote a server as a primary
//
func (v *View) HandlePrimaryFailure() {
	v.Primary = v.Backup
	v.Backup = ""
	v.Viewnum += 1
}

//
// demote a server as a backup
//
func (v *View) BackupFailure() {
	v.Backup = ""
	v.Viewnum += 1
}

type ViewServer struct {
	mu               sync.Mutex
	l                net.Listener
	dead             bool
	me               string
	availableServers AvailableServers
	currentView      View
	primaryHasAcked  bool
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {
	// Add server to list of available servers
	if args.Viewnum == 0 {
		vs.availableServers.Add(args.Me)
	} else if len(args.Me) > 0 {
		vs.availableServers.Update(args.Me)
	}

	server, isServerFound := vs.availableServers.FindServer(args.Me)

	if isServerFound && vs.currentView.Primary == args.Me && (args.Viewnum == vs.currentView.Viewnum) {
		vs.primaryHasAcked = true
		server.Ack()
	}

	reply.View = vs.currentView

	return nil
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {
	reply.View = vs.currentView
	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	// Break out early if we have no servers
	if vs.availableServers.Len() < 1 {
		return
	}

	vs.CleanupDeadServers()

	// If we don't have a primary promote one from the available servers
	if len(vs.currentView.Primary) < 1 && vs.primaryHasAcked {
		newPrimary, _ := vs.availableServers.GetIdle()
		newPrimary.Status = Primary
		vs.currentView.PromotePrimary(*newPrimary)
		vs.primaryHasAcked = false
	} else if len(vs.currentView.Backup) < 1 && vs.primaryHasAcked {
		newBackup, isIdler := vs.availableServers.GetIdle()
		if isIdler {
			newBackup.Status = Backup
			vs.currentView.PromoteBackup(*newBackup)
		}
		vs.primaryHasAcked = false
	}
}

//
// Determines if there are dead servers and then demotes them if needed
//
func (vs *ViewServer) CleanupDeadServers() {
	deadServers, isDead := vs.availableServers.KillDeadServers()

	if isDead && vs.primaryHasAcked {
		server := deadServers[0]
		if vs.IsServerPrimary(server) {
			vs.currentView.HandlePrimaryFailure()
			vs.primaryHasAcked = false
		} else if vs.IsServerBackup(server) {
			vs.currentView.BackupFailure()
			vs.primaryHasAcked = false
		}
	}
}

//
// Checks if the server is the current primary server
//
func (vs *ViewServer) IsServerPrimary(server Server) bool {
	return server.Me == vs.currentView.Primary
}

//
// Checks if the server is the current backup
//
func (vs *ViewServer) IsServerBackup(server Server) bool {
	return server.Me == vs.currentView.Backup
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
	vs.availableServers = AvailableServers{}
	vs.currentView = View{Viewnum: 0}
	vs.primaryHasAcked = true

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
