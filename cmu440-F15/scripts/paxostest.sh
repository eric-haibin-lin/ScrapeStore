#!/bin/bash

if [ -z $GOPATH ]; then
    echo "FAIL: GOPATH environment variable is not set"
    exit 1
fi

if [ -n "$(go version | grep 'darwin/amd64')" ]; then    
    GOOS="darwin_amd64"
elif [ -n "$(go version | grep 'linux/amd64')" ]; then
    GOOS="linux_amd64"
else
    echo "FAIL: only 64-bit Mac OS X and Linux operating systems are supported"
    exit 1
fi

# Build srunner
go install github.com/cmu440-F15/paxosapp/runners/srunner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Build the student's paxos node implementation.
# Exit immediately if there was a compile-time error.
go install github.com/cmu440-F15/paxosapp/runners/prunner
if [ $? -ne 0 ]; then
   echo "FAIL: code does not compile"
   exit $?
fi

# Pick random ports between [10000, 20000).
SLAVE_PORT0=$(((RANDOM % 10000) + 10000))
SLAVE_PORT1=$(((RANDOM % 10000) + 10000))
SLAVE_PORT2=$(((RANDOM % 10000) + 10000))
SLAVE_PORT3=$(((RANDOM % 10000) + 10000))
SLAVE_PORT4=$(((RANDOM % 10000) + 10000))
SLAVE_PORT5=$(((RANDOM % 10000) + 10000))
SLAVE_NODE=$GOPATH/bin/srunner
ALL_SLAVE_PORTS="${SLAVE_PORT0},${SLAVE_PORT1},${SLAVE_PORT2},${SLAVE_PORT3},${SLAVE_PORT4},${SLAVE_PORT5}"

echo $ALL_SLAVE_PORTS
# Start slave node.
${SLAVE_NODE} -hostport="localhost:$SLAVE_PORT0" -id=0 2 &
SLAVE_NODE_PID0=$!
sleep 1

${SLAVE_NODE} -hostport="localhost:$SLAVE_PORT1" -id=1 2 &
SLAVE_NODE_PID1=$!
sleep 1

${SLAVE_NODE} -hostport="localhost:$SLAVE_PORT2" -id=2 2 &
SLAVE_NODE_PID2=$!
sleep 1

${SLAVE_NODE} -hostport="localhost:$SLAVE_PORT3" -id=3 2 &
SLAVE_NODE_PID3=$!
sleep 1

${SLAVE_NODE} -hostport="localhost:$SLAVE_PORT4" -id=4 2 &
SLAVE_NODE_PID4=$!
sleep 1

${SLAVE_NODE} -hostport="localhost:$SLAVE_PORT5" -id=5 2 &
SLAVE_NODE_PID5=$!
sleep 1


# Pick random ports between [10000, 20000).
NODE_PORT0=$(((RANDOM % 10000) + 10000))
NODE_PORT1=$(((RANDOM % 10000) + 10000))
NODE_PORT2=$(((RANDOM % 10000) + 10000))

echo "localhost:$NODE_PORT0,localhost:$NODE_PORT1,localhost:$NODE_PORT2" > paxosnodesport.txt
#TESTER_PORT=$(((RANDOM % 10000) + 10000))
#PROXY_PORT=$(((RANDOM & 10000) + 10000))
#PAXOS_TEST=$GOPATH/bin/paxostest
PAXOS_NODE=$GOPATH/bin/prunner
ALL_PORTS="${NODE_PORT0},${NODE_PORT1},${NODE_PORT2}"

##################################################

# Start paxos node.
${PAXOS_NODE} -ports=${ALL_PORTS} -slaveports=${ALL_SLAVE_PORTS} -N=3 -id=0 -numslaves=6 2 &
PAXOS_NODE_PID0=$!
sleep 1

${PAXOS_NODE} -ports=${ALL_PORTS} -slaveports=${ALL_SLAVE_PORTS} -N=3 -id=1 -numslaves=6 2 &
PAXOS_NODE_PID1=$!
sleep 1

${PAXOS_NODE} -ports=${ALL_PORTS} -slaveports=${ALL_SLAVE_PORTS} -N=3 -id=2 -numslaves=6 2 &
PAXOS_NODE_PID2=$!
sleep 1

echo "Sleeping for 10 seconds..."
sleep 10

while [[ true ]]; do
  #statements
done
# kill -9 ${SLAVE_NODE_PID0}
# kill -9 ${SLAVE_NODE_PID1}
# kill -9 ${SLAVE_NODE_PID2}
# kill -9 ${SLAVE_NODE_PID3}
# kill -9 ${SLAVE_NODE_PID4}
# kill -9 ${SLAVE_NODE_PID5}

# Kill paxos node.
# kill -9 ${PAXOS_NODE_PID0}
# kill -9 ${PAXOS_NODE_PID1}
# kill -9 ${PAXOS_NODE_PID2}

# exit
