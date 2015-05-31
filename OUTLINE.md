# Course Outline

This is the course outline that we followed for the course:

## Week 0 (4/27)
  - Reading: [MapReduce](readings/mapreduce.pdf)
  - Lecture Notes: [Introduction](http://css.csail.mit.edu/6.824/2014/notes/l01.txt)
  - Homework: None

## Week 1 (5/4)
  - Reading: [MapReduce](readings/mapreduce.pdf)
  - Lecture Notes: [At-Most-Once RPC](http://css.csail.mit.edu/6.824/2014/notes/l02.txt), [handout](http://css.csail.mit.edu/6.824/2014/notes/l02-rpc-mutex.go)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 1](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-i-word-count)
  - Suggested Readings:
    - [Golang Getting Started](https://golang.org/doc/install)
    - [A tour of Go](https://tour.golang.org/welcome/1)
    - [Effective Go](https://golang.org/doc/effective_go.html)

## Week 2 (5/11)
  - Reading: [MapReduce](readings/mapreduce.pdf)
    - Question: "How soon after it receives the first file of intermediate data can a reduce worker start calling the application's Reduce function? Explain your answer."
  - Lecture Notes: [Fault Tolerance: primary/backup replication](http://css.csail.mit.edu/6.824/2014/notes/l03.txt)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 2](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-ii-distributing-mapreduce-jobs)
  - Suggested Readings:
    - [Effective Go: Concurrency](https://golang.org/doc/effective_go.html#concurrency)
    - [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)

## Week 3 (5/18)
  - Reading: [HyperVisor](readings/bressoud-hypervisor.pdf)
    - Question: "What do you like best about Go? Why? Would you want to change anything in the language? If so, what and why?"
  - Homework: [Part 3](labs/lab-1.md#part-iii-handling-worker-failures)
  - Suggested Readings:
    - [Effective Go: Concurrency](https://golang.org/doc/effective_go.html#concurrency)
    - [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)

## Week 4 (5/27)
  - Reading: [HyperVisor](readings/bressoud-hypervisor.pdf)
    - Question: "Suppose that instead of connecting both the primary and backup to the same disk, we connected them to separate disks with identical copies of the data? Would this work? What else might we have to worry about, and what are some things that could go wrong?"
  - Lecture Notes: [Fault Tolerance: primary/backup replication](http://css.csail.mit.edu/6.824/2014/notes/l03.txt)
  - Homework: [Lab 2](labs/lab-2.md) | [Part A](labs/lab-2.md#part-a-the-viewservice)

## Week 5 (6/1)
  - Reading: [Flat Datacenter Storage](readings/fds.pdf)
    - Question: "Suppose tractserver T1 is temporarily unreachable due to a network problem, so the metadata server drops T1 from the TLT. Then the network problem goes away, but for a while the metadata server is not aware that T1's status has changed. During this time could T1 serve client requests to read and write tracts that it stores? If yes, give an example of how this could happen. If no, explain what mechanism(s) prevent this from happening."
  - Lecture Notes: [Flat Datacenter Storage case study](http://css.csail.mit.edu/6.824/2014/notes/l04.txt)
  - Homework: [Lab 2](labs/lab-2.md) | Begin [Part B](labs/lab-2.md#part-b-the-primarybackup-keyvalue-service) (Do steps 1-4)

## Week 6 (6/8)
  - Reading: [Paxos](readings/paxos-simple.pdf)
  - Lecture Notes: [Paxos](http://css.csail.mit.edu/6.824/2014/notes/l05-paxos.txt)
    - Question: "Suppose that the acceptors are A, B, and C. A and B are also proposers. How does Paxos ensure that the following sequence of events can't happen? What actually happens, and which value is ultimately chosen?
    A sends prepare requests with proposal number 1, and gets responses from A, B, and C.
    A sends accept(1, "foo") to A and C and gets responses from both. Because a majority accepted, A thinks that "foo" has been chosen. However, A crashes before sending an accept to B.
    B sends prepare messages with proposal number 2, and gets responses from B and C.
    B sends accept(2, "bar") messages to B and C and gets responses from both, so B thinks that "bar" has been chosen."

  - Homework: [Lab 2](labs/lab-2.md) | [Part B](labs/lab-2.md#part-b-the-primarybackup-keyvalue-service)

## Week 7 (6/15)
  - Reading: [Replication in the Harp File System](readings/bliskov-harp.pdf)
    - Question: "Figures 5-1, 5-2, and 5-3 show that Harp often finishes benchmarks faster than a conventional non-replicated NFS server. This may be surprising, since you might expect Harp to do strictly more work than a conventional NFS server (for example, Harp must manage the replication). Why is Harp often faster? Will all NFS operations be faster with Harp than on a conventional NFS server, or just some of them? Which?"
  - Homework: [Lab 3](labs/lab-3.md) | Part A

## Week 8 (6/22)
  - Reading: [Epaxos: There Is More Consensus in Egalitarian Parliaments](readings/epaxos.pdf)
    - Question: "When will an EPaxos replica R execute a particular command C? Think about when commands are committed, command interference, read operations, etc."
  - Homework: Begin [Lab 3](labs/lab-3.md) | Part B

## Week 9 (6/29)
  - Reading: [Spanner](readings/spanner.pdf)
    - Question: "Suppose a Spanner server's TT.now() returns correct information, but the uncertainty is large. For example, suppose the absolute time is 10:15:30, and TT.now() returns the interval [10:15:20,10:15:40]. That interval is correct in that it contains the absolute time, but the error bound is 10 seconds. See Section 3 for an explanation TT.now(). What bad effect will a large error bound have on Spanner's operation? Give a specific example."
  - Homework: [Lab 3](labs/lab-3.md) | Part B

## Week 10 (7/6)
  - Reading: Transaction Chains (2013)
    - Question:  In the simple straw-man, both fetch and modify operations are placed in the log and signed. Suppose an alternate design that only signs and logs modify operations. Does this allow a malicious server to break fetch-modify consistency or fork consistency? Why or why not?
  - Homework: [Lab 4](labs/lab-4.md) | Part A

## Week 11 (7/13)
  - Reading: Shared Virtual Memory (1986) && Treadmarks (1994)
    - Question: Memory Coherence in Shared Virtual Systems ivy-code.txt is a version of the code in Section 3.1 with some clarifications and bug fixes. The manager part of the WriteServer sends out invalidate messages, and waits for confirmation messages indicating that the invalidates have been received and processed. Suppose the manager send out invalidates, but did not wait for confirmations. Describe a scenario in which lack of the confirmation would cause the system to behave incorrectly. You should assume that the network delivers all messages, and that none of the computers fail.
  - Homework: Begin [Lab 4](labs/lab-4.md) | Part B

## Week 12 (7/20)
  - Reading: Spark (2012)
  - Question: What applications can Spark support well that MapReduce/Hadoop cannot support?
  - Homework: Continue [Lab 4](labs/lab-4.md) | Part B

## Week 13 (7/27)
  - Reading: Ficus (1994) && Bayou (1995)
    - Question About Ficus: Imagine a situation like the paper's Figure 1, but in which only Site A updates file Foo. What should Ficus do in that case when the partition is merged? Explain how Ficus could tell the difference between the situation in which both Site A and Site B update Foo, and the situation in which only Site A updates Foo.
    - Question Bayou: Suppose we build a distributed filesystem using Bayou, and the system has a copy operation. Initially, file A contains "foo" and file B contains "bar". On one node, a user copies file A to file B, overwriting the old contents of B. On another node, a user copies file B to file A. After both operations are committed, we want both files to contain "foo" or for both files to contain "bar". Sketch a dependency check and merge procedure for the copy operation that makes this work. How does Bayou ensure that all the nodes agree about whether A and B contain "foo" or "bar"?
  - Homework: Finish [Lab 4](labs/lab-4.md) | Part B

## Week 14 (8/3)
  - Reading: PNUTS (2008) && Dynamo (2007)
    - Question PNUTS: Briefly explain why it is (or isn't) okay to use relaxed consistency for social applications (see Section 4). Does PNUTS handle the type of problem presented by Example 1 in Section 1, and if so, how?
    - Question Dynamo: Suppose Dynamo server S1 is perfectly healthy with a working network connection. By mistake, an administrator instructs server S2 to remove S1 using the mechanisms described in 4.8.1 and 4.9. It takes a while for the membership change to propagate from S2 to the rest of the system (including S1), so for a while some clients and servers will think that S1 is still part of the system. Will Dynamo operate correctly in this situation? Why, or why not?
  - Homework: Final Project

## Week 15 (8/10)
  - Reading: Kademilia (2002) && Trackerless Bittorent (2008)
    - Question Kademilia: Consider a Kademlia-based key-value store with a million users, with non-mutable keys: once a key is published, it will not be modified. The k/v store experiences a network partition into two roughly equal partitions A and B for 1.5 hours.

    X is a very popular key. Would nodes in both A and B likely be able to access X's value (1) during the partition? (2) 10 minutes after the network is joined? (3) 25 hours after the network is joined?

    (optional) Would your answer change if X was an un-popular key?
  - Homework: Final Project

## Week 16 (8/17)
  - Reading: Akkamai && Hubspot blog post
  - Homework: Present Final Project

## Week 17 (8/24)
  - Reading: Bitcoin && AnalogicFS
    - Question Bitcoin: Try to buy something with Bitcoin. It may help to cooperate with some 6.824 class-mates, and it may help to start a few days early. If you decide to give up, that's OK. Briefly describe your experience.
    - Question AnalogicFS: In many ways, this experiences paper raises more questions than it answers. Please answer one of the following questions, taking into consideration the rich history of AnalogicFS and the spirit in which the paper was written:

    a) The analysis of A* search shown in Figure 1 claims to be an introspective visualization of the AnalogicFS methodology; however, not all decisions are depicted in the figure. In particular, if I <= P, what should be the next node explored such that all assumptions in Section 2 still hold? Show your work.

    b) Despite the authors' claims in the introduction that AnalogicFS was developed to study SCSI disks (and their interaction with lambda calculus), the experimental setup detailed in Section 4.1 involves decommissioned Gameboys instead, which use cartridge-based, Flash-like memory. If the authors had used actual SCSI disks during the experiments, how exactly might have their results changed quantitatively?

    c) AnalogicFS shows rather unstable multicast algorithm popularity (Figure 5), especially compared with some of the previous systems we've read about in 6.824. Give an example of another system that would have a more steady measurement of popularity pages, especially in the range of 0.1-0.4 decibels of bandwidth.

    d) For his 6.824 project, Ben Bitdiddle chose to build a variant of Lab 4 that faithfully emulates the constant expected seek time across LISP machines, as AnalogicFS does. Upon implementation, however, he immediately ran into the need to cap the value size to 400 nm, rather than 676 nm. Explain what assumptions made for the AnalogicFS implementation do not hold true for Lab 4, and why that changes the maximum value size.
  - Homework: Present Final Project Cont'd

Additional Papers:
[Google Photon (2013)](http://research.google.com/pubs/pub41318.html)

# Idealized Outline - WIP

## Week 0
  - Reading: [MapReduce](readings/mapreduce.pdf)

  - Homework [Lab 1](labs/lab-1.md) | [Part 1](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-i-word-count)
  - Suggested Readings:
    - [Golang Getting Started](https://golang.org/doc/install)
    - [A tour of Go](https://tour.golang.org/welcome/1)
    - [Effective Go](https://golang.org/doc/effective_go.html)

## Week 1
  - Reading: [MapReduce](readings/mapreduce.pdf)
  - Homework: [Lab 1](labs/lab-1.md) | [Part 2](https://github.com/keathley/6.824/blob/master/labs/lab-1.md#part-ii-distributing-mapreduce-jobs) and [Part 3](labs/lab-1.md#part-iii-handling-worker-failures)
  - Suggested Readings:
    - [Effective Go: Concurrency](https://golang.org/doc/effective_go.html#concurrency)
    - [Go Concurrency Patterns: Pipelines and cancellation](https://blog.golang.org/pipelines)

## Week 2
  - Reading: [HyperVisor](readings/bressoud-hypervisor.pdf)
  - Homework: [Lab 2] | [Part A]

## Week 3
  - Reading: [Flat Datacenter Storage](readings/fsd.pdf)
  - Homework: [Lab 2] | [Part B]
