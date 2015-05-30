# Course Outline

This is the course outline that we followed for the course:

* Week 0 (4/27): [MapReduce](readings/mapreduce.pdf)
  - Lecture Notes: [Introduction](http://css.csail.mit.edu/6.824/2014/notes/l01.txt)
  - Homework: None
* Week 1 (5/4): [MapReduce](readings/mapreduce.pdf)
  - Lecture Notes: [At-Most-Once RPC](http://css.csail.mit.edu/6.824/2014/notes/l02.txt), [handout](http://css.csail.mit.edu/6.824/2014/notes/l02-rpc-mutex.go)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 1](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-i-word-count)
  - Suggested Readings:
    - [Golang Getting Started](https://golang.org/doc/install)
    - [A tour of Go](https://tour.golang.org/welcome/1)
    - [Effective Go](https://golang.org/doc/effective_go.html)
* Week 2 (5/11): [MapReduce](readings/mapreduce.pdf)
  - Question: "How soon after it receives the first file of intermediate data can a reduce worker start calling the application's Reduce function? Explain your answer."
  - Lecture Notes: [Fault Tolerance: primary/backup replication](http://css.csail.mit.edu/6.824/2014/notes/l03.txt)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 2](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-ii-distributing-mapreduce-jobs)
  - Suggested Readings:
    - [Effective Go: Concurrency](https://golang.org/doc/effective_go.html#concurrency)
    - [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)
* Week 3 (5/18): [HyperVisor](readings/bressoud-hypervisor.pdf)
  - Question: "What do you like best about Go? Why? Would you want to change anything in the language? If so, what and why?"
  - Homework: [Part 3](labs/lab-1.md#part-iii-handling-worker-failures)
  - Suggested Readings:
    - [Effective Go: Concurrency](https://golang.org/doc/effective_go.html#concurrency)
    - [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)
* Week 4 (5/27): [HyperVisor](readings/bressoud-hypervisor.pdf)
  - Question: "Suppose that instead of connecting both the primary and backup to the same disk, we connected them to separate disks with identical copies of the data? Would this work? What else might we have to worry about, and what are some things that could go wrong?"
  - Lecture Notes: [Fault Tolerance: primary/backup replication](http://css.csail.mit.edu/6.824/2014/notes/l03.txt)
  - Homework: [Lab 2](labs/lab-2.md) | [Part A](labs/lab-2.md#part-a-the-viewservice)
* Week 5 (6/1): [Flat Datacenter Storage](readings/fds.pdf)
  - Question: "Suppose tractserver T1 is temporarily unreachable due to a network problem, so the metadata server drops T1 from the TLT. Then the network problem goes away, but for a while the metadata server is not aware that T1's status has changed. During this time could T1 serve client requests to read and write tracts that it stores? If yes, give an example of how this could happen. If no, explain what mechanism(s) prevent this from happening."
  - Lecture Notes: [Flat Datacenter Storage case study](http://css.csail.mit.edu/6.824/2014/notes/l04.txt)
  - Homework: [Lab 2](labs/lab-2.md) | Begin [Part B](labs/lab-2.md#part-b-the-primarybackup-keyvalue-service) (Do steps 1-4)
* Week 6 (6/8): [Paxos](readings/paxos-simple.pdf)
  - Lecture Notes: [Paxos](http://css.csail.mit.edu/6.824/2014/notes/l05-paxos.txt)
  - Question: "Suppose that the acceptors are A, B, and C. A and B are also proposers. How does Paxos ensure that the following sequence of events can't happen? What actually happens, and which value is ultimately chosen?

A sends prepare requests with proposal number 1, and gets responses from A, B, and C.
A sends accept(1, "foo") to A and C and gets responses from both. Because a majority accepted, A thinks that "foo" has been chosen. However, A crashes before sending an accept to B.
B sends prepare messages with proposal number 2, and gets responses from B and C.
B sends accept(2, "bar") messages to B and C and gets responses from both, so B thinks that "bar" has been chosen."
  - Homework: [Lab 2](labs/lab-2.md) | [Part B](labs/lab-2.md#part-b-the-primarybackup-keyvalue-service)
* Week 7 (6/15): [Replication in the Harp File System](readings/bliskov-harp.pdf)
  - Question: "Figures 5-1, 5-2, and 5-3 show that Harp often finishes benchmarks faster than a conventional non-replicated NFS server. This may be surprising, since you might expect Harp to do strictly more work than a conventional NFS server (for example, Harp must manage the replication). Why is Harp often faster? Will all NFS operations be faster with Harp than on a conventional NFS server, or just some of them? Which?"
  - Homework: [Lab 3](labs/lab-3.md) | Part A
* Week 8 (6/22): [Epaxos: There Is More Consensus in Egalitarian Parliaments](readings/epaxos.pdf)
  - Question: "When will an EPaxos replica R execute a particular command C? Think about when commands are committed, command interference, read operations, etc."
  - Homework: Begin [Lab 3](labs/lab-3.md) | Part B
* Week 9 (6/29): [Spanner](readings/spanner.pdf)
  - Question: "Suppose a Spanner server's TT.now() returns correct information, but the uncertainty is large. For example, suppose the absolute time is 10:15:30, and TT.now() returns the interval [10:15:20,10:15:40]. That interval is correct in that it contains the absolute time, but the error bound is 10 seconds. See Section 3 for an explanation TT.now(). What bad effect will a large error bound have on Spanner's operation? Give a specific example."
  - Homework: [Lab 3](labs/lab-3.md) | Part B
* Week 10 (7/6): Transaction Chains (2013)
  - Homework: [Lab 4](labs/lab-4.md) | Part A
* Week 11 (7/13): Shared Virtual Memory (1986) && Treadmarks (1994)
  - Homework: Begin [Lab 4](labs/lab-4.md) | Part B
* Week 12 (7/20): Spark (2012)
  - Homework: Continue [Lab 4](labs/lab-4.md) | Part B
* Week 13 (7/27): Ficus (1994) && Bayou (1995)
  - Homework: Finish [Lab 4](labs/lab-4.md) | Part B
* Week 14 (8/3): PNUTS (2008) && Dynamo (2007)
  - Homework: Final Project
* Week 15 (8/10): Kademilia (2002) && Trackerless Bittorent (2008)
  - Homework: Final Project
* Week 16 (8/17): Akkamai && Hubspot blog post
  - Homework: Present Final Project
* Week 17 (8/24): Bitcoin && AnalogicFS
  - Homework: Present Final Project Cont'd

Additional Papers:
[Google Photon (2013)](http://research.google.com/pubs/pub41318.html)

# Idealized Outline

* Week 0: [MapReduce](readings/mapreduce.pdf)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 1](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-i-word-count)
  - Suggested Readings:
    - [Golang Getting Started](https://golang.org/doc/install)
    - [A tour of Go](https://tour.golang.org/welcome/1)
    - [Effective Go](https://golang.org/doc/effective_go.html)
* Week 1: [MapReduce](readings/mapreduce.pdf)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 2](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-ii-distributing-mapreduce-jobs) and [Part 3](labs/lab-1.md#part-iii-handling-worker-failures)
  - Suggested Readings:
    - [Effective Go: Concurrency](https://golang.org/doc/effective_go.html#concurrency)
    - [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)
* Week 2: [HyperVisor](readings/bressoud-hypervisor.pdf)
  - Homework: [Lab 2] | [Part A]
* Week 3: [Flat Datacenter Storage](readings/fsd.pdf)
  - Homework: [Lab 2] | [Part B]
