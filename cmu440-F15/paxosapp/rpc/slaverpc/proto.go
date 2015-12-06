// This file contains constants and arguments used to perform RPCs between
// Slave nodes and Master nodes.

package slaverpc

type GetArgs struct {
	Key string
}

type GetReply struct {
	Value interface{}
}

type AppendArgs struct {
	Key string
	Value interface{}
}

type AppendReply struct {
	//Nothing to reply here 
}