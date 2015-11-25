package paxos

import (
	"errors"
	"fmt"
	"github.com/cmu440-F15/paxosapp/rpc/paxosrpc"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type paxosNode struct {
	// TODO: implement this!
	myHostPort string
	numNodes   int
	srvId      int
	hostMap    map[int]string

	/* The main key-value store */
	valuesMap map[string]interface{}

	/* Note - acceptedValuesMap and acceptedSeqNumMap should be used hand-in-hand */

	/* Temporary map which should be populated only when accept is called and
	should be cleared when commit is called */
	acceptedValuesMap map[string]interface{}

	/* Temporary map which should be populated only when accept is called and
	should be cleared when commit is called */
	acceptedSeqNumMap map[string]int

	/* Again, a temporary map which should be populated when prepare is called */
	maxSeqNumSoFar map[string]int

	valuesMapLock *sync.Mutex
}

// NewPaxosNode creates a new PaxosNode. This function should return only when
// all nodes have joined the ring, and should return a non-nil error if the node
// could not be started in spite of dialing the other nodes numRetries times.
//
// hostMap is a map from node IDs to their hostports, numNodes is the number
// of nodes in the ring, replace is a flag which indicates whether this node
// is a replacement for a node which failed.
func NewPaxosNode(myHostPort string, hostMap map[int]string, numNodes, srvId, numRetries int, replace bool) (PaxosNode, error) {
	fmt.Println("myhostport is ", myHostPort, "Numnodes is ", numNodes, "srvid is ", srvId)

	var a paxosrpc.RemotePaxosNode

	node := paxosNode{}

	node.srvId = srvId
	node.numNodes = numNodes
	node.myHostPort = myHostPort
	node.hostMap = make(map[int]string)

	node.valuesMap = make(map[string]interface{})

	node.acceptedValuesMap = make(map[string]interface{})
	node.acceptedSeqNumMap = make(map[string]int)

	node.maxSeqNumSoFar = make(map[string]int)

	node.valuesMapLock = &sync.Mutex{}

	for k, v := range hostMap {
		node.hostMap[k] = v
	}

	listener, err := net.Listen("tcp", myHostPort)
	if err != nil {
		return nil, err
	}

	a = &node

	err = rpc.RegisterName("PaxosNode", paxosrpc.Wrap(a))
	if err != nil {
		return nil, err
	}

	rpc.HandleHTTP()
	go http.Serve(listener, nil)

	for _, v := range hostMap {
		_, err := rpc.DialHTTP("tcp", v)

		cntr := 0

		if err != nil {
			for {
				fmt.Println(myHostPort, " couldn't dial", v, ". Trying again.")

				cntr = cntr + 1

				if cntr == numRetries {
					fmt.Println("Couldn't connect even after all retries.", myHostPort, " aborting.")
					return nil, errors.New("Couldn't dial a node")
				}
				time.Sleep(1 * time.Second)
				_, err := rpc.DialHTTP("tcp", v)
				if err == nil {
					break
				}
			}
		}
		fmt.Println(myHostPort, " dialed ", v, " successfully")
	}

	return a, nil
}

func (pn *paxosNode) GetNextProposalNumber(args *paxosrpc.ProposalNumberArgs, reply *paxosrpc.ProposalNumberReply) error {
	fmt.Println("GetNextProposalNumber invoked on ", pn.srvId)
	return errors.New("not implemented")
}

func prepare(pn *paxosNode, hostport string, key string, seqnum int, preparechan chan paxosrpc.PrepareReply) {
	dialer, err := rpc.DialHTTP("tcp", hostport)

	if err != nil {
		fmt.Println("ERROR: Couldn't Dial prepare on ", hostport)
		return
	}

	args := paxosrpc.PrepareArgs{}
	args.Key = key
	args.N = seqnum

	reply := paxosrpc.PrepareReply{}

	err = dialer.Call("PaxosNode.RecvPrepare", &args, &reply)

	if err != nil {
		fmt.Println("RPC RecvPrepare failed!")
		return
	}

	preparechan <- reply
}

func accept(pn *paxosNode, hostport string, value interface{}, key string, seqnum int, acceptchan chan paxosrpc.AcceptReply) {
	dialer, err := rpc.DialHTTP("tcp", hostport)

	if err != nil {
		fmt.Println("ERROR: Couldn't Dial accept on ", hostport)
		return
	}

	args := paxosrpc.AcceptArgs{}
	args.Key = key
	args.N = seqnum
	args.V = value

	reply := paxosrpc.AcceptReply{}

	err = dialer.Call("PaxosNode.RecvAccept", &args, &reply)

	if err != nil {
		fmt.Println("RPC RecvAccept failed!")
		return
	}

	acceptchan <- reply
}

func commit(pn *paxosNode, hostport string, value interface{}, key string, commitchan chan int) {
	dialer, err := rpc.DialHTTP("tcp", hostport)

	if err != nil {
		fmt.Println("ERROR: Couldn't Dial commit on ", hostport)
		return
	}

	args := paxosrpc.CommitArgs{}
	args.Key = key
	args.V = value

	reply := paxosrpc.CommitReply{}

	err = dialer.Call("PaxosNode.RecvCommit", &args, &reply)

	if err != nil {
		fmt.Println("RPC RecvCommit failed!")
		return
	}

	commitchan <- 1
}
func (pn *paxosNode) Propose(args *paxosrpc.ProposeArgs, reply *paxosrpc.ProposeReply) error {
	preparechan := make(chan paxosrpc.PrepareReply)
	acceptchan := make(chan paxosrpc.AcceptReply)
	commitchan := make(chan int)

	fmt.Println("In Propose of ", pn.srvId)

	fmt.Println("Key is ", args.Key, ", V is ", args.V, " and N is ", args.N)

	for _, v := range pn.hostMap {
		fmt.Println("Will call Prepare on ", v)
		go prepare(pn, v, args.Key, args.N, preparechan)
	}

	okcount := 0

	max_n := 0
	max_v := args.V

	for i := 0; i < pn.numNodes; i++ {
		ret := <-preparechan
		if ret.Status == paxosrpc.OK {
			okcount++
			if ret.N_a != 0 && ret.N_a > max_n {
				max_n = ret.N_a
				max_v = ret.V_a
			}
		}
	}

	if !(okcount >= ((pn.numNodes / 2) + 1)) {
		return errors.New("Didn't get a majority in prepare phase")
	}

	var valueToPropose interface{}
	if max_n != 0 { //someone suggested a different value
		valueToPropose = args.V
	} else {
		valueToPropose = max_v
	}

	for _, v := range pn.hostMap {
		fmt.Println("Will call Accept on ", v)
		go accept(pn, v, valueToPropose, args.Key, args.N, acceptchan)
	}

	okcount = 0

	for i := 0; i < pn.numNodes; i++ {
		ret := <-acceptchan
		if ret.Status == paxosrpc.OK {
			okcount++
		}
	}

	if !(okcount >= ((pn.numNodes / 2) + 1)) {
		return errors.New("Didn't get a majority in accept phase")
	}

	for _, v := range pn.hostMap {
		fmt.Println("Will call Commit on ", v)
		go commit(pn, v, valueToPropose, args.Key, commitchan)
	}

	for i := 0; i < pn.numNodes; i++ {
		_ = <-commitchan
	}

	reply.V = valueToPropose

	pn.valuesMapLock.Lock()
	pn.valuesMap[args.Key] = valueToPropose
	pn.valuesMapLock.Unlock()

	return nil
}

func (pn *paxosNode) GetValue(args *paxosrpc.GetValueArgs, reply *paxosrpc.GetValueReply) error {
	pn.valuesMapLock.Lock()
	val, ok := pn.valuesMap[args.Key]

	if ok {
		reply.V = val
		reply.Status = paxosrpc.KeyFound
		pn.valuesMapLock.Unlock()
		return nil
	}
	pn.valuesMapLock.Unlock()

	reply.Status = paxosrpc.KeyNotFound
	return nil
}

func (pn *paxosNode) RecvPrepare(args *paxosrpc.PrepareArgs, reply *paxosrpc.PrepareReply) error {
	fmt.Println("In RecvPrepare of ", pn.myHostPort)
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvAccept(args *paxosrpc.AcceptArgs, reply *paxosrpc.AcceptReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvCommit(args *paxosrpc.CommitArgs, reply *paxosrpc.CommitReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvReplaceServer(args *paxosrpc.ReplaceServerArgs, reply *paxosrpc.ReplaceServerReply) error {
	return errors.New("not implemented")
}

func (pn *paxosNode) RecvReplaceCatchup(args *paxosrpc.ReplaceCatchupArgs, reply *paxosrpc.ReplaceCatchupReply) error {
	return errors.New("not implemented")
}

/*

// This file contains constants and arguments used to perform RPCs between
// two Paxos nodes. DO NOT MODIFY!

package paxosrpc

// Status represents the status of a RPC's reply.
type Status int
type Lookup int

const (
	OK     Status = iota + 1 // Paxos replied OK
	Reject                   // Paxos rejected the message
)

const (
	KeyFound    Lookup = iota + 1 // GetValue key found
	KeyNotFound                   // GetValue key not found
)

type ProposalNumberArgs struct {
	Key string
}

type ProposalNumberReply struct {
	N int
}

type ProposeArgs struct {
	N   int // Proposal number
	Key string
	V   interface{} // Value for the Key
}

type ProposeReply struct {
	V interface{} // Value that was actually committed for that key
}

type GetValueArgs struct {
	Key string
}

type GetValueReply struct {
	V      interface{}
	Status Lookup
}

type PrepareArgs struct {
	Key string
	N   int
}

type PrepareReply struct {
	Status Status
	N_a    int         // Highest proposal number accepted
	V_a    interface{} // Corresponding value
}

type AcceptArgs struct {
	Key string
	N   int
	V   interface{}
}

type AcceptReply struct {
	Status Status
}

type CommitArgs struct {
	Key string
	V   interface{}
}

type CommitReply struct {
	// No content, no reply necessary
}

type ReplaceServerArgs struct {
	SrvID    int // Server being replaced
	Hostport string
}

type ReplaceServerReply struct {
	// No content necessary
}

type ReplaceCatchupArgs struct {
	// No content necessary
}

type ReplaceCatchupReply struct {
	Data []byte
}
*/
