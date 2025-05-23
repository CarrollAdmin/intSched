# Prerequisites

```
OS: Centos 7 (Full Installation Guide) or any other Linux Distro if Kubernetes and NFS are present
    
Docker: 20.10+

Kubernetes: 1.18 -> 1.21

Docker Registry        (For multi-node setup)

NFS                    (For multi-node setup)
```

# intSched Deployment

# Initialize Setup Parameters

## Initialize NFS Path (MANDATORY for multi-node setups)
In order to share executables and datasets across all the Virtual Slurm Cluster's Nodes we have to initialize the NFS_PATH file 
with the path of the NFS filesystem. 

If single node Kubernetes is available then use a system path.

```
sudo nano Build/NFS_PATH
```

eg.
```
/var/nfs_share_dir/
```


## Initialize Docker Registry IP (OPTIONAL for Local Builds ONLY)
If you wish to modify intSched or Virtual Slurm Cluster Dockerfiles and build your own modified images then a Docker Registry is needed.
In order to pull and push Docker Images we have to initialize Docker Registry IP in REGISTRY file with IP of the Docker Registry we installed earlier.

```
sudo nano Build/REGISTRY
```

eg.
```
REGISTRY_IP:5000
``` 


# intSched Scheduler Deployemt

In this section we are going to configure and deploy intSched.

## Deploy intSched 

Run:
```
(cd intSched/ && make install-dockerhub)
```

## If you wish to build your own modified image with different scheduling parameters

To configure intSched modify the intSched initialization command to the "intSched/intSched_init_command.sh" script.

intSched Configuration options:
```
    # Usage of intSched command line options:

    #    -Allow_MPI_Colocation string
    #            -Allow_MPI_Colocation 1 (0 disbale MPI JOB Colocation) (default "1")
    #    -Allow_Spreading_of_Tasks string
    #            -Allow_Spreading 1 ("0" Max Loaded Node Selection, "1" Least Loaded Node Selection) (default "1")
    #    -Allow_Task_Colocation string
    #            -Allow_Task_Colocation 1 (0 disable Colocation) (default "1")
    #    -Exclude_Nodelist string
    #            -Exclude_Nodelist <Node-1,Node-2..,Node-n> (eg. node1,node2,nodeN)
    #    -Include_Nodelist string
    #            -Include_Nodelist <Node-1,Node-2..,Node-n> (eg. node1,node2,nodeN)
    #    -Kube_Config_Path string
    #            Path to Kubernetes Configuration file (default "/home/master/.kube/Config")
    #    -Max_Node_Capacity string
    #            -Max_Node_Capacity 1 (100% Node Capacity) (default "1")
    #    -Namespaces string
    #            -Namespaces <namespace-1,namespace-2..,namespace-n> (default "default")
    #    -PID_Scale string
    #            -PID_Scale 5 (PID Scaler larger values accelerate resource scaling while losing accuracy) (default "3")
    #    -Scheduler_Name string
    #            -Scheduler Name (default "intSched")
    # 
```
Initialization Command Example:
```
     ./intSched -Namespaces namespace1,default -Include_Nodelist node1,node2,node3,node4
```

Deploy Modified intSched 

After we have configured the "intSched/intSched_init_command.sh" script we are going to deploy intSched.

Run:
```
(cd intSched/ && make install)
```

The "make install" option builds and pushes the container to the local docker registry and then deploys the modified image.
    
Check if intSched Container is up and running.
    
Run:
``` 
kubectl get pods -o wide  
```

Output:
```  
intSched-7f94f85b5b-9cxrt   1/1     Running   0          65s   10.244.1.3   minikube-m02   <none>           <none>
```
    
    
# Virtual Slurm Cluster Deployment (Ethernet)

If only Ethernet adapters are available to the Cluster's Nodes
we can run Slurm HPC workloads by using Virtual Slurm Clusters (Ethernet version) 
that uses the classic TCP-UDP protocol for the containers communication. 

This version runs on every classic Data-Center hardware setup 
as it does not require any specialized hardware.

## Virtual Slurm Cluster(Ethernet) Installation

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install-dockerhub)
```



## Deploy Modified Virtual-Cluster-Ethernet

If you wish to build your own modified image with different pre-installed libraries and dependencies modify the Dockerfile.
After you have configured the Dockerfile.

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install)
```

The "make install" option builds and pushes the container to the local docker registry and then deploys the modified image.  


# Prometheus Deployment

We are going to deploy Prometheus with a Custom-Metrics-Server in order for intSched to pull custom performance metrics from the applications running.

## Deploy Prometheus Operator

Run:
```
(cd Build/Prometheus/prometheus-operator/ && make install-dockerhub)
```

## Deploy Prometheus

Run:
```
(cd Build/Prometheus/prometheus/ && make install-dockerhub)
```

## Deploy Custom Metric Server

Run:
```
(cd Build/Prometheus/metrics-server/ && make install-dockerhub)
```  

# HPC Demos

# Single Node Demo (Only Kubernetes needed)

In this demo we are going to deploy intSched + Virtual Slurm Clusters (Ethernet) on a single Node Kubernetes environment.

This demo is for demonstration purposes.

## Prerequisites (IMPORTANT)

A working Kubernetes __1.18 -> 1.21__ installation.

## Configure and Deploy intSched (Step 1)

### Install intSched

Run:
```
(cd intSched/ && make install-dockerhub)
```


Check if intSched Container is up and running.
    
Run:
``` 
kubectl get pods -o wide  
```

Output:
```    
intSched-XXX-XXX   1/1     Running   0          65s   10.244.1.3   minikube-m02   <none>           <none>
```

## Virtual Slurm Cluster Deployment (Ethernet) (Step 2)

This version runs on every classic Data-Center hardware setup 
as it does not require any specialized hardware.

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install-dockerhub)
```


## MPI Demo Single Node (NAS Parallel Benchmarks) (Step 3)

In this example we are going to run some MPI benchmarks from the NAS Parallel Benchmarks suite, 
using the Virtual Slurm Cluster (Execution Environment) + intSched (Scheduler) concept.

### Connect to a Virtual Slurm Cluster's Container
In order to use a Virtual Slurm Cluster we have to connect to one of its containers.

Check if Virtual Slurm Cluster's Containers are up and running.
Run:
``` 
kubectl get pods -o wide  
```

This command will return a list with all the running Kubernetes pods inside our namespace.
The Virtual Slurm Cluster's containers are named intSched-slurm-ethernet-XXX or intSched-slurm-infiniband-XXX.
We choose one the containers in order to connect with a terminal:.
    
Output:
``` 
intSched-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.1.57    worker-node1   <none>           <none>
intSched-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.0.112   master-node    <none>           <none>
```

This command will open a terminal inside the Virtual Slurm Cluster's selected container:  
    
Run:
``` 
kubectl exec -it intSched-slurm-ethernet-XXX bash                             
```

### Download "NAS Parallel Benchmarks" 

Now we are going to download and build the "NAS Parallel Benchmarks" 
inside the /nfs/ directory.

Inside the container Run:
``` 
git clone  https://github.com/wzzhang-HIT/NAS-Parallel-Benchmark.git /nfs/NAS-Parallel-Benchmark &&
cd /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI && 
mv config/make.def.template config/make.def &&
mv config/suite.def.template config/suite.def &&
sed -i 's/MPIF77 = f77/MPIF77 = mpifort/' config/make.def &&
sed -i 's/S/C/' config/suite.def &&
sed -i 's/1/4/' config/suite.def &&
mkdir bin
```

### Build the Benchmarks

In order to build the benchmarks 

Run:
``` 
make suite
```

Now the binaries are located under the "/nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin" directory

### Run a Benchmark

In order to start an MPI benchmark with Slurm 

Run:
``` 
srun --mpi=pmix -N 1 -n 4 /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin/cg.C.4
```

The above command is going to run the cg.C.4 benchmark on 1 Node with 4 processes in total.

# Multi Node Minikube Demo (Easy to follow, all dependencies included by Minikube)


In this demo we are going to deploy intSched + Virtual Slurm Clusters (Ethernet) on a 2 Node Minikube Kubernetes environment.

This demo is for demonstration purposes.


## Prerequisites

A working Minikube installation.

See below how to set up Minikube:

https://minikube.sigs.k8s.io/docs/start/

Create a directory in order to be accessible by all Virtual Slurm Cluster Pods
```
sudo mkdir /var/nfs_share_dir/
sudo chmod -R 755 /var/nfs_share_dir
```



Add Minikube Docker Registry to the REGISTRY file in order to be used by the upcoming deployments
```
echo "127.0.0.1:5000" > REGISTRY
```

## Start the Minikube environment using Docker (Step 1)
```
minikube start --driver=docker  --kubernetes-version=v1.19.7 --nodes 2 --cpus 4 --mount-string="/var/nfs_share_dir/:/var/nfs_share_dir/" --mount
```

## Deploy intSched (Step 2)
In this section we are going to deploy intSched


Run:
```
(cd intSched/ && make install-dockerhub)
```


Check if intSched Container is up and running.
    
Run:
``` 
kubectl get pods -o wide  
```

Output:   
```  
intSched-7f94f85b5b-9cxrt   1/1     Running   0          65s   10.244.1.3   minikube-m02   <none>           <none>    
```

## Virtual Slurm Cluster Deployment (Ethernet) (Step 3)

If only Ethernet adapters are available to the Cluster's Nodes
we can run Slurm HPC workloads by using Virtual Slurm Clusters (Ethernet version) 
that uses the classic TCP-UDP protocol for the containers communication.

This version runs on every classic Data-Center hardware setup 
as it does not require any specialized hardware.

Run:
``` 
(cd Build/Virtual-Cluster-Slurm-Ethernet && make install-dockerhub)
```

## MPI Demo Single Node (NAS Parallel Benchmarks) (Step 4)

In this example we are going to run some MPI benchmarks from the NAS Parallel Benchmarks suite, 
using the Virtual Slurm Cluster (Execution Environment) + intSched (Scheduler) concept.

### Connect to a Virtual Slurm Cluster's Container
In order to use a Virtual Slurm Cluster we have to connect to one of its containers.

Check if Virtual Slurm Cluster's Containers are up and running.
    
Run:
    
``` 
kubectl get pods -o wide  
```

This command will return a list with all the running Kubernetes pods inside our namespace.
The Virtual Slurm Cluster's containers are named intSched-slurm-ethernet-XXX or intSched-slurm-infiniband-XXX.

Output:
    
``` 
intSched-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.1.57    worker-node1   <none>           <none>
intSched-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.0.112   master-node    <none>           <none>   
```


​    
We choose one the containers in order to connect with a terminal:.
​    
This command will open a terminal inside the Virtual Slurm Cluster's selected container:  
​    
Run:
``` 
kubectl exec -it intSched-slurm-ethernet-XXX bash                         
```

### Download "NAS Parallel Benchmarks" 

Now we are going to download and build the "NAS Parallel Benchmarks" 
inside the /nfs/ directory.

Inside the container Run:
``` 
git clone  https://github.com/wzzhang-HIT/NAS-Parallel-Benchmark.git /nfs/NAS-Parallel-Benchmark &&
cd /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI && 
mv config/make.def.template config/make.def &&
mv config/suite.def.template config/suite.def &&
sed -i 's/MPIF77 = f77/MPIF77 = mpifort/' config/make.def &&
sed -i 's/S/C/' config/suite.def &&
sed -i 's/1/4/' config/suite.def &&
mkdir bin
```

### Build the Benchmarks

In order to build the benchmarks 

Run:
``` 
make suite
```

Now the binaries are located under the "/nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin" directory

### Run a Benchmark

In order to start an MPI benchmark with Slurm 

Run:
``` 
srun --mpi=pmix -N 2 -n 4 /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin/cg.C.4
```

The above command is going to run the cg.C.4 benchmark on 2 Nodes with 4 processes in total.


# Complete Demo

In this example we are going to run some MPI benchmarks from the NAS Parallel Benchmarks suite, 
using the Virtual Slurm Cluster (Execution Environment) + intSched (Scheduler) concept.

## Connect to a Virtual Slurm Cluster's Container (Step 1)
In order to use a Virtual Slurm Cluster we have to connect to one of its containers.

Run:
``` 
kubectl get pods -o wide
```

Output:
``` 
intSched-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.1.57    worker-node1   <none>           <none>
intSched-slurm-ethernet-XXX           1/1     Running   0          2m37s   10.244.0.112   master-node    <none>           <none>
```

This command will return a list with all the running Kubernetes pods inside our namespace.
The Virtual Slurm Cluster's containers are named intSched-slurm-ethernet-XXX or intSched-slurm-infiniband-XXX.    
    
We choose one of the containers in order to connect with a terminal:

Run:
``` 
kubectl exec -it intSched-slurm-ethernet-XXX bash                         
```

This command will open a terminal inside the Virtual Slurm Cluster's selected container.       
    
## Download "NAS Parallel Benchmarks"  (Step 2)

Now we are going to download and build the "NAS Parallel Benchmarks" 
inside the NFS directory that is shared among the containers at the "/nfs" directory.

Inside the container 
    
Run:
``` 
git clone  https://github.com/wzzhang-HIT/NAS-Parallel-Benchmark.git /nfs/NAS-Parallel-Benchmark &&
cd /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI && 
mv config/make.def.template config/make.def &&
mv config/suite.def.template config/suite.def &&
sed -i 's/MPIF77 = f77/MPIF77 = mpifort/' config/make.def &&
mkdir bin
```

## Configure the Benchmarks  (Step 3)

Modify the config/suite.def in order to choose the parameters 
that the benchmarks are going to be compiled with.

Run:
``` 
nano config/suite.def
```

Example configuration:
``` 
# config/suite.def
# This file is used to build several benchmarks with a single command.
# Typing "make suite" in the main directory will build all the benchmarks
# specified in this file.
# Each line of this file contains a benchmark name, class, and number
# of nodes. The name is one of "cg", "is", "ep", mg", "ft", "sp", "bt",
# "lu", and "dt".
# The class is one of "S", "W", "A", "B", "C", "D", and "E"
# (except that no classes C, D and E for DT, and no class E for IS).
# The number of nodes must be a legal number for a particular
# benchmark. The utility which parses this file is primitive, so
# formatting is inflexible. Separate name/class/number by tabs.
# Comments start with "#" as the first character on a line.
# No blank lines.
# The following example builds 4 processor sample sizes of all benchmarks.

ep      C       4
cg      C       4

```


## Build the Benchmarks  (Step 4)

In order to build the benchmarks 

Run:
``` 
make suite
```

Now the binaries are located under the "/nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin" directory

## Run a Benchmark  (Step 5)

In order to start an MPI benchmark with Slurm 

Run:
``` 
srun --mpi=pmix -N "NUM_OF_NODES" -n "NUM_OF_PROCESSES" /nfs/NAS-Parallel-Benchmark/NPB3.3-MPI/bin/"REPLACE_BENCHMARK_BINARY"
```

The above command is going to run the "REPLACE_BENCHMARK_BINARY" benchmark binary on "NUM_OF_NODES" nodes with "NUM_OF_PROCESSES" processes in total.
