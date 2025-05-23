#!/bin/bash

# Default run.
if [ $# -eq 0 ]; then
    ssh-keygen -A
    /usr/sbin/sshd -D
    exit 0
fi

LABEL=${1:-slurmd}
SLOTS=${2:-2}
NAMESPACE=${3:-slurm}

# Configure ssh and create a key for this host.
echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config
ssh-keygen -A

# Wait for all the pods to become ready.
while [[ $(kubectl get pods -n=${NAMESPACE} -l app=${LABEL} -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}' | xargs -n 1 | sort -u) != "True" ]]; do
    echo "Waiting for pod(s)..."
    sleep 1
done

MASTER_NODE=$(kubectl get nodes --selector='node-role.kubernetes.io/master' -o jsonpath='{.items[0].metadata.name}')
MASTER_POD=$(kubectl get pods -n=${NAMESPACE} -l app=${LABEL} -o jsonpath='{range .items[?(@.spec.nodeName=="'$MASTER_NODE'")]}{.metadata.name}')
echo "Master pod is: $MASTER_POD"

# Set up user key and known hosts file.
PODS=(`kubectl get pods -n=${NAMESPACE} -l app=${LABEL} -o jsonpath='{.items[*].metadata.name}'`)
if [ `hostname` == ${MASTER_POD} ]; then
    echo "Creating ssh keys and other files..."
    ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa
    cat /root/.ssh/id_rsa.pub > /root/.ssh/authorized_keys
    for POD in ${PODS[@]}; do
        KEY=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c 'cat /etc/ssh/ssh_host_ecdsa_key.pub'`)
        ADDRESS=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c 'getent hosts | grep 'slurm' | cut -d " " -f1'`)
        echo "$ADDRESS ${KEY[0]} ${KEY[1]}" >> /root/.ssh/known_hosts
        echo "$ADDRESS slots=${SLOTS}" >> /root/hostfile
    done

    echo "Configuring Slurm..."
    if [ -f /private/slurm.conf ]; then
        echo "Starting with configuration found in /private..."
        cp -f /private/slurm.conf /etc/slurm.conf
    else
        echo "Creating configuration file with defaults..."
        cat > /etc/slurm.conf << EOF
AuthType=auth/none
CredType=cred/none
MpiDefault=none
ProctrackType=proctrack/linuxproc
ReturnToService=1
SlurmctldPidFile=/var/run/slurmctld.pid
SlurmctldPort=6817
SlurmdPidFile=/var/run/slurmd.pid
SlurmdPort=6818
SlurmdSpoolDir=/var/spool/slurmd
SlurmUser=root
StateSaveLocation=/var/spool
SwitchType=switch/none
TaskPlugin=task/none
PropagatePrioProcess=2
InactiveLimit=0
KillWait=30
MinJobAge=300
SlurmctldTimeout=120
SlurmdTimeout=300
Waittime=0
SchedulerType=sched/backfill
SelectType=select/cons_tres
SelectTypeParameters=CR_Core
AccountingStorageType=accounting_storage/none
AccountingStoreJobComment=YES
ClusterName=cluster
JobCompType=jobcomp/none
JobAcctGatherFrequency=30
JobAcctGatherType=jobacct_gather/none
SlurmctldDebug=info
SlurmdDebug=info
GresTypes=gpu
EOF
    fi

    echo "Adding hosts to configuration file..."
    echo "SlurmctldHost="`hostname`"("`hostname -i`")" >> /etc/slurm.conf
    NODES=""
    for POD in ${PODS[@]}; do
        CPUS=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c "lscpu | grep -E '^CPU\(' | cut -d':' -f2 | tr -d '[:space:]'"`)
        SOCKETS=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c "lscpu | grep -E '^Socket' | cut -d':' -f2 | tr -d '[:space:]'"`)
        CORES_PER_SOCKET=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c "lscpu | grep -E '^Core' | cut -d':' -f2 | tr -d '[:space:]'"`)
        THREADS_PER_CORE=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c "lscpu | grep -E '^Thread' | cut -d':' -f2 | tr -d '[:space:]'"`)
        ADDRESS=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c 'getent hosts | grep 'slurm' | cut -d " " -f1'`)
        # GPU=(`1`) # hardcode
        NODES=$NODES","$ADDRESS
        echo "NodeName=$ADDRESS CPUs=$CPUS SocketsPerBoard=$SOCKETS CoresPerSocket=$CORES_PER_SOCKET ThreadsPerCore=$THREADS_PER_CORE State=UNKNOWN" >> /etc/slurm.conf
    done
    NODES=`echo $NODES | cut -c2-`
    echo "PartitionName=mpi Nodes=$NODES Default=YES MaxTime=INFINITE State=UP" >> /etc/slurm.conf

    echo "Copying over files to other pod(s)..."
    declare -a FILES=("/root/.ssh/id_rsa" "/root/.ssh/id_rsa.pub" "/root/.ssh/authorized_keys" "/root/.ssh/known_hosts" "/root/hostfile" "/etc/slurm.conf")
    for POD in ${PODS[@]}; do
        if [ $POD != ${MASTER_POD} ]; then
            kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c 'mkdir -p /root/.ssh && chmod 700 /root/.ssh'
            for FILE in ${FILES[@]}; do
                kubectl cp -n=${NAMESPACE} $FILE $POD:$FILE
            done
        fi
    done

    echo "Starting Slurm ..."
    for POD in ${PODS[@]}; do
        if [ $POD != ${MASTER_POD} ]; then
            ADDRESS=(`kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c 'getent hosts | grep 'slurm' | cut -d " " -f1'`)
            kubectl exec -n=${NAMESPACE} $POD -- /bin/bash -c "slurmd -N $ADDRESS"
        fi
    done
    slurmd -N `hostname -i`
    slurmctld
fi

# Start the ssh service
echo "Starting ssh..."
/usr/sbin/sshd -D