package main

//
// toy RPC library using Go channels
//

import "io"
import "fmt"
import "encoding/binary"
import "sync"

type Request struct {
  Xid int64
  ProcNum int32
  Arg int32
}

type Reply struct {
  Xid int64
  Res int32
}

type RequestChan struct {
  request Request
  done chan int32
}

type ToyClient struct {
  conn io.ReadWriteCloser    // connection to server
  requestchan chan RequestChan
  replychan chan Reply
  done sync.WaitGroup
}

func MakeToyClient(conn io.ReadWriteCloser) *ToyClient {
  tc := &ToyClient{}
  tc.conn = conn
  tc.requestchan = make(chan RequestChan)
  tc.replychan = make(chan Reply)
  go tc.Listener()
  go tc.Multiplexer()
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
  done := make(chan int32) // for tc.Listener()
  tc.requestchan <- RequestChan{Request{int64(0), procNum, arg}, done}
  reply := <- done // wait for reply via tc.Listener()
  return reply
}

// Client goroutine that multiplexes request and demultiplexes replies
func (tc *ToyClient) Multiplexer() {
  var xid int64
  xid = 1
  pending := map[int64]chan int32{}
  for {
    select {
    case req := <- tc.requestchan:
      pending[xid] = req.done
      tc.WriteRequest(&Request{xid, req.request.ProcNum, req.request.Arg})
      xid++
    case rep := <- tc.replychan:
      pending[rep.Xid] <- rep.Res
      delete(pending, rep.Xid)
    }
  }
}

// client goroutine that waits for a reply and sends it to multiplexer
func (tc *ToyClient) Listener() {
  for {
    reply := tc.ReadReply()
    tc.replychan <- *reply
  }
}

type ToyServer struct {
  conn io.ReadWriteCloser // connection from client
  replychan chan Reply
  handlers map[int32]func(int32)int32 // procedures
}

func MakeToyServer(conn io.ReadWriteCloser, h map[int32](func(int32)int32)) *ToyServer {
  ts := &ToyServer{}
  ts.conn = conn
  ts.handlers = h
  ts.replychan = make(chan Reply)
  go ts.Dispatcher()
  go ts.Writer()
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

// listen for client requests and dispatch them to the right handler
func (ts *ToyServer) Dispatcher() {
  for {
    req := ts.ReadRequest()
    fn, ok := ts.handlers[req.ProcNum]
    go func() {
      reply := Reply{req.Xid, 0}
      if ok {
        reply.Res = fn(req.Arg)
      }
      ts.replychan <- reply  // Sends reply through writer
    }()
  }
}

func (ts *ToyServer) Writer() {
  for {
    select {
    case rep := <- ts.replychan:
      ts.WriteReply(&rep)
    }
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
  handlers := map[int32](func(int32)int32){}
  handlers[22] = func(a int32) int32 { 
    return a+1 
  }
  tc := MakeToyClient(cp)
  MakeToyServer(sp, handlers)
  tc.done.Add(3)
  go test(tc, 1)
  go test(tc, 2)
  go test(tc, 3)
  tc.done.Wait()
}
