package pbservice

import "github.com/keathley/6.824/viewservice"
import "net/rpc"

import "fmt"

// You'll probably need to uncomment these:
import "strconv"
import "time"
import "crypto/rand"
import "math/big"

type Clerk struct {
	vs *viewservice.Clerk
	// Your declarations here
	me     string
	getnum int
	putnum int
}

func MakeClerk(vshost string, me string) *Clerk {
	ck := new(Clerk)
	ck.vs = viewservice.MakeClerk(me, vshost)
	// Your ck.* initializations here
	ck.me = strconv.FormatInt(nrand(), 10)
	ck.getnum = 1
	ck.putnum = 1
	return ck
}
func nrand() int64 {
	max := big.NewInt(int64(10000000))
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, rpcname string,
	args interface{}, reply interface{}) bool {
	c, errx := rpc.Dial("unix", srv)
	if errx != nil {
		return false
	}
	defer c.Close()

	err := c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}

//
// fetch a key's value from the current primary;
// if they key has never been set, return "".
// Get() must keep trying until it either the
// primary replies with the value or the primary
// says the key doesn't exist (has never been Put().
//
func (ck *Clerk) Get(key string) string {
	args := &GetArgs{key, ck.getnum, ck.me}
	var reply GetReply
	ck.doGet(args, &reply)
	ck.getnum++
	return reply.Value
}

func (ck *Clerk) doGet(args *GetArgs, reply *GetReply) {
	for {
		//		DPrintf("Trying Get...\n")
		view, ok := ck.vs.Get()
		if ok {
			ok = call(view.Primary, "PBServer.Get", args, reply)
			//			DPrintf("Reply: %s\n", reply)
		}
		if reply.Err == OK || reply.Err == ErrNoKey {
			break
		}
		time.Sleep(viewservice.PingInterval)
	}
}

func (ck *Clerk) doPut(args *PutArgs, reply *PutReply) {
	for {
		//		DPrintf("Trying Put...\n")
		view, ok := ck.vs.Get()
		if ok {
			call(view.Primary, "PBServer.Put", args, reply)
			//			DPrintf("Reply: %s\n", reply)
		}
		if reply.Err == OK {
			break
		}
		time.Sleep(viewservice.PingInterval)
	}
}

//
// tell the primary to update key's value.
// must keep trying until it succeeds.
//
func (ck *Clerk) PutExt(key string, value string, dohash bool) string {
	args := &PutArgs{key, value, dohash, ck.putnum, ck.me}
	var reply PutReply
	ck.doPut(args, &reply)
	ck.putnum++
	return reply.PreviousValue
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutExt(key, value, false)
}
func (ck *Clerk) PutHash(key string, value string) string {
	v := ck.PutExt(key, value, true)
	return v
}
