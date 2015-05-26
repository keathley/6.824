package viewservice

import "net"
import "net/rpc"
import "log"
import "time"
import "sync"
import "fmt"
import "os"

type Server struct {
	Me       string
	PingTime time.Time
}

type AvailableServers struct {
	servers []Server
}

func (as *AvailableServers) UpdateServer(name string) bool {
	fmt.Println("Updating Server:", name)

	server, isFound := as.FindServer(name)
	server.PingTime = time.Now()

	if !isFound {
		fmt.Println("Server was found")
		server.Me = name
		as.servers = append(as.servers, server)
	}

	return false
}

func (as *AvailableServers) FindServer(name string) (Server, bool) {

	for _, server := range as.servers {
		if server.Me == name {
			return server, true
		}
	}

	return Server{}, false
}

func (as *AvailableServers) Len() int {
	return len(as.servers)
}

func (as *AvailableServers) GetIdle() Server {
	fmt.Println("Getting idle server")
	return as.servers[0]
}

type ViewServer struct {
	mu               sync.Mutex
	l                net.Listener
	dead             bool
	me               string
	availableServers AvailableServers
	currentView      View
}

//
// server Ping RPC handler.
//
func (vs *ViewServer) Ping(args *PingArgs, reply *PingReply) error {
	// Add server to list of available servers
	if len(args.Me) > 0 {
		vs.availableServers.UpdateServer(args.Me)
	}

	// Reply with current view
	reply.View = vs.currentView

	return nil
}

//
// server Get() RPC handler.
//
func (vs *ViewServer) Get(args *GetArgs, reply *GetReply) error {

	fmt.Println("GET")
	fmt.Println("GET View:", vs.currentView)

	reply.View = vs.currentView

	// Your code here.

	return nil
}

//
// tick() is called once per PingInterval; it should notice
// if servers have died or recovered, and change the view
// accordingly.
//
func (vs *ViewServer) tick() {
	// Break out early if we have no servers
	fmt.Println("SERVER COUNT:", vs.availableServers.Len())
	if vs.availableServers.Len() < 1 {

		return
	}

	// If we don't have a primary promote one from the available servers
	// vs.current
	if len(vs.currentView.Primary) < 1 {
		primary := vs.availableServers.GetIdle()
		vs.currentView = vs.buildNewView(primary.Me, "")
	}

	fmt.Println("currentView", vs.currentView)
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

func (vs *ViewServer) buildNewView(primary string, backup string) View {
	newView := View{
		Viewnum: vs.currentView.Viewnum + 1,
		Primary: primary,
		Backup:  backup,
	}

	fmt.Println("new View", newView)
	return newView
}

func StartServer(me string) *ViewServer {
	vs := new(ViewServer)
	vs.me = me
	vs.availableServers = AvailableServers{}
	vs.currentView = View{Viewnum: 0}

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
