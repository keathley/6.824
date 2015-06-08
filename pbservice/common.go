package pbservice

import "hash/fnv"
import "fmt"

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongServer = "ErrWrongServer"
	NotBackup      = "~not backup~"
)

type Err string

type PutArgs struct {
	Key     string
	Value   string
	DoHash  bool // For PutHash
	Version int  // first custom field
	Client  string
}

type GenericReply struct {
	Err Err
}

type PutReply struct {
	GenericReply
	PreviousValue string // For PutHash
}

type GetArgs struct {
	Key     string
	Version int // first custom field
	Client  string
}

type GetReply struct {
	GenericReply
	Value string
}

type BackupArgs struct {
	Data       map[string]string
	RequestLog map[string]*ClientState
}

type BackupReply struct {
	Err Err
}

type State struct {
	Version int
	Value   string
}

func (state State) String() string {
	return fmt.Sprintf("{Version: %d, Value: '%s'}", state.Version, state.Value)
}

type ClientState struct {
	Get *State
	Put *State
}

func (state ClientState) String() string {
	return fmt.Sprintf("ClientState:{Get: %s\tPut: %s}", state.Get, state.Put)
}

// Your RPC definitions here.

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
