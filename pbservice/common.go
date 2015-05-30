package pbservice

import "hash/fnv"

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongServer = "ErrWrongServer"
)

type Err string

type PutArgs struct {
	Key     string
	Value   string
	DoHash  bool // For PutHash
	Version uint // first custom field
	Client  string
	// Field names must start with capital letters,
	// otherwise RPC will break.
}

type PutReply struct {
	Err           Err
	PreviousValue string // For PutHash
}

type GetArgs struct {
	Key     string
	Version uint // first custom field
	Client  string
}

type GetReply struct {
	Err   Err
	Value string
}

type BackupArgs struct {
	Data       map[string]string
	RequestLog map[string]ClientState
}

type BackupReply struct {
	Err Err
}

type ClientState struct {
	lastGet       uint
	lastGetResult string
	lastPut       uint
	lastPutResult string
}

// Your RPC definitions here.

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
