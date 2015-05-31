package main

//
// toy RPC library with mutexes and condition variables
//

import "io"
import "fmt"
import "sync"
import "encoding/binary"

type Request struct {
  Xid int64
  ProcNum int32
  Arg int32
}

type Reply struct {
  Xid int64
  Res int32
}

type State struct {
  cond *sync.Cond
  reply *Reply
}

type ToyClient struct {
  mu sync.Mutex
  conn io.ReadWriteCloser    // connection to server
  xid int64                  // next unique request #
  pending map[int64]*State
  done sync.WaitGroup
}

func MakeToyClient(conn io.ReadWriteCloser) *ToyClient {
  tc := &ToyClient{}
  tc.conn = conn
  tc.pending = map[int64]*State{}
  tc.xid = 1
  go tc.Listener()
  return tc
}

func (tc *ToyClient) WriteRequest(req *Request) {
  err := binary.Write(tc.conn, binary.LittleEndian, req)
  if err != nil {
		fmt.Println("binary.Write failed:", err)
  }
}

func (tc *ToyClient) ReadReply() *Reply {
  reply := Reply{}
  binary.Read(tc.conn, binary.LittleEndian, &reply)
  return &reply
}

// client application uses Call() to make an RPC.
func (tc *ToyClient) Call(procNum int32, arg int32) int32 {
  tc.mu.Lock()
  defer tc.mu.Unlock()

  xid := tc.xid // allocate a unique xid
  tc.xid++
  tc.pending[xid] = &State{sync.NewCond(&tc.mu), nil}
  req := &Request{xid,procNum, arg}
  tc.WriteRequest(req) // send to server

  for tc.pending[xid].reply == nil {
    tc.pending[xid].cond.Wait()
  }

  r := tc.pending[xid].reply
  delete(tc.pending, xid)
  return r.Res
}

// client goroutine that waits for a reply
func (tc *ToyClient) Listener() {
  for {
    reply := tc.ReadReply()
    tc.mu.Lock()
    entry, ok := tc.pending[reply.Xid]
    if ok {
      entry.reply = reply;
      entry.cond.Signal()
    }
    tc.mu.Unlock()
  }
}

type ToyServer struct {
  mu sync.Mutex
  conn io.ReadWriteCloser // connection from client
  handlers map[int32]func(int32)int32 // procedures
}

func MakeToyServer(conn io.ReadWriteCloser) *ToyServer {
  ts := &ToyServer{}
  ts.conn = conn
  ts.handlers = map[int32](func(int32)int32){}
  go ts.Dispatcher()
  return ts
}

func (ts *ToyServer) WriteReply(reply *Reply) {
  err := binary.Write(ts.conn, binary.LittleEndian, reply)
  if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
}

func (ts *ToyServer) ReadRequest() *Request {
  req := Request{};
  err := binary.Read(ts.conn, binary.LittleEndian, &req)
  if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
  return &req
}

//
// listen for client requests, dispatch them to the right handler, send back
// replies.
//
func (ts *ToyServer) Dispatcher() {
  for {
    req := ts.ReadRequest()
    ts.mu.Lock()
    fn, ok := ts.handlers[req.ProcNum]
    ts.mu.Unlock()
    go func() {
      reply := Reply{req.Xid, 0}
      if ok {
        reply.Res = fn(req.Arg)
      }
      ts.mu.Lock()
      ts.WriteReply(&reply)
      ts.mu.Unlock()
    }()
  }
}

type Pair struct {
  r *io.PipeReader
  w *io.PipeWriter
}
func (p Pair) Read(data []byte) (int, error) {
  return p.r.Read(data)
}
func (p Pair) Write(data []byte) (int, error) {
  return p.w.Write(data)
}
func (p Pair) Close() error {
  p.r.Close()
  return p.w.Close()
}

func test(tc *ToyClient, arg int32) {
  for i := 0;  i < 1000; i++ {
    reply := tc.Call(22, arg)
    fmt.Printf("Call(22, %d) -> %v\n", arg, reply)
  }
  tc.done.Done()
}

func main() {
  r1, w1 := io.Pipe()
  r2, w2 := io.Pipe()
  cp := Pair{r : r1, w : w2}
  sp := Pair{r : r2, w : w1}
  tc := MakeToyClient(cp)
  ts := MakeToyServer(sp)
  ts.handlers[22] = func(a int32) int32 { 
    return a+1 
  }
  tc.done.Add(3)
  go test(tc, 1)
  go test(tc, 2)
  go test(tc, 3)
  tc.done.Wait()
}
