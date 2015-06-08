<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
<link rel="StyleSheet" href="../style.css" type="text/css">
<title>6.824 Lab 5: Final Project</title>
</head>

## [6.824](../index.html) - Spring 2014

# 6.824 Lab 5: Final Project

| **Piazza idea discussions due:** | Thursday, March 20, 2014 |
| **Proposals due:** | Friday, April 4, 2014 |
| **Code and write-up due:** | Friday, May 9, 2014 (11:59pm) |
| **Presentations:** | Tuesday, May 13, 2014 and Thursday, May 15, 2014 (in class) |

* * *

## Introduction

In this lab, you will work on a final project of your own choice. Unlike in previous labs, you will work in groups of 3-4 for the final project. You will be required to turn in both your code and a short write-up describing the design and implementation of your project, and to make a short in-class presentation about your work. We will post your write-up and code on the web site after the end of the semester, unless you explicitly talk to us about why you want to keep yours confidential.

The primary requirement is that your project be something interesting. Your project should also have something to do with distributed systems, but that's relatively easy, and it's much more important for your project to be interesting.

We encourage students to choose any project idea that you might think is interesting. If you are not sure, we provide you with two kinds of potential projects. First, we present a reasonably well-defined starting point for your project -- basically, a default project. Second, we give a list of half-baked ideas that we think could turn into an interesting project, but we haven't given them too much thought.

We encourage final projects that leverage multiple classes you might be taking, or that involve other research or projects you are already working on (e.g., MEng and AUP projects).

## Deliverables

There are four concrete steps to the final project, as follows:

Form a group.Decide on the project you would like to work on, and post short summary of your idea (one to two paragraphs) on Piazza; use the <tt>hw5</tt> tag/folder on Piazza.

Discuss ideas with others in comments on their Piazza posting. Use these postings to help find other students interested in similar ideas for forming a group. Course staff will provide feedback on project ideas on Piazza; if you'd like more detailed feedback, come chat with us in person.

Project proposal.Discuss your proposed idea with course staff over the next week, before the proposal deadline, to flesh out the exact problem you will be addressing, how you will go about doing it, and what tools you might need in the process. By the proposal deadline, you must submit a one-to-two-page proposal describing: your **group members** list, **the problem** you want to address, **how you plan to address it** , and what are you proposing to **specifically design and implement**.

Submit your proposal to [https://ydmao.scripts.mit.edu:444/6.824/handin.py](https://ydmao.scripts.mit.edu:444/6.824/handin.py)

Write-up and code.Write a document describing the design and implementation of your project, and turn it in along with your project's code by the final deadline. The document should be about 3 pages of text that helps us understand what problem you solved, and what your code does. The code and writeups will be posted online after the end of the semester.

Project presentation.Prepare a short in-class presentation about the work that you have done for your final project. We will provide a projector that you can use to demonstrate your project.

## A well-defined starting point

Your goal for the default project is to build a persistent, fault-tolerant, high-performance key/value store, using lab 4 as a starting point. This means:

- Your servers should store data persistently on disk. You must handle the following cases: 
  - Your replicated service should be able to recover from a single server crashing and restarting with its disk contents intact, by catching up the state of that server. 
  - Your replicated service should be able to recover from a single server crashing and losing its disk contents. When the server comes back up, your service should bring the server up-to-date. 
  - Your replicated service should be able to recover from a complete crash of all servers. (Of course, until a majority of them come back up, your service will not be able to handle client requests.) 

Note that making your service persistent implies that you will need to keep some Paxos state around on disk, in addition to your key/value state.

- You should write test cases that check whether your service operates correctly under a wide range of failure scenarios. A good starting point are the test cases from labs 3 and 4, but you will need to extend them to deal with persistence.
- Your service should be able to handle large amounts of data (e.g., hundreds of gigabytes per server), particularly when servers temporarily or permanently fail. Have a look at RAMCloud, Flat Datacenter Storage, Petal, and FAB for ideas.
- Your service should achieve high performance in terms of client requests per second, and in terms of client request latency. In particular, you should make two optimizations to your Paxos-based protocol: avoid 2 round-trips per agreement (e.g., by having a server issue Prepare messages ahead of time), and avoid dueling leaders under high client load (e.g., by using a designated leader like Multi-Paxos, using a striping approach like Mencius, or by using EPaxos). 

You can take a look at the source code for EPaxos [here](https://github.com/efficient/epaxos). Feel free to integrate it into your final project, although doing so may be quite non-trivial, and you may be better off implementing the above two optimizations yourself.

- You should write benchmarks to evaluate the throughput and latency of client operations on your service.
- You should identify the bottlenecks that are responsible for the throughput and latency results you observe.

## Half-baked project ideas

Here's a list of ideas to get you started thinking -- but you should feel free to pursue your own ideas.

- Make the state synchronization protocol (DDP) in [Meteor](http://www.meteor.com/) more efficient (e.g., send fewer bytes between server and client) and more fault-tolerant (e.g., a client should be able to tolerate server failures, as long as enough servers remain live).
- Build a fault-tolerant file service; on the client side, you could use FUSE to run your own client code, or you could have clients talk NFS to your server, as in Harp.
- Build a better fault-tolerant peer-to-peer tracker for BitTorrent.
- Build a system for making Node.js applications fault-tolerant, perhaps in the style of Hypervisor Fault Tolerance.
- Improve the [Roost](https://roost.mit.edu/) Javascript Zephyr client by replicating the backend to make it fault-tolerant.
- Use [Mylar](http://css.csail.mit.edu/mylar/) to build a secure webmail system with end-to-end encryption, contact lists, usable public key infrastructure, etc.
- Add cross-shard atomic transactions to Lab 4, using two-phase commit and/or snapshots.
- Build a system with asynchronous replication (like Dynamo or Ficus or Bayou). Perhaps add stronger consistency (as in COPS or Walter or Lynx).
- Build a file synchronizer (like [Unison](http://www.cis.upenn.edu/~bcpierce/unison/) or [Tra](http://swtch.com/tra/)).
- Build a [distributed shared memory](http://www.cdf.toronto.edu/~csc469h/fall/handouts/nitzberg91.pdf) (DSM) system, so that you can run multi-threaded shared memory parallel programs on a cluster of machines, using paging to give the appearance of real shared memory. When a thread tries to access a page that's on another machine, the page fault will give the DSM system a chance to fetch the page over the network from whatever machine currently stores.
- Build a distributed RAID in the style of FAB. Maybe you can get standard operating systems to talk to you network virtual disk using iSCSI or Linux's NBD (network block device).
- Build a coherent caching system for use by web sites (a bit like memcached), perhaps along the lines of [TxCache](http://drkp.net/papers/txcache-osdi10.pdf).
- Build a distributed cooperative web cache, perhaps along the lines of [Firecoral](https://www.usenix.org/legacy/events/iptps09/tech/full_papers/terrace/terrace_html/) or [Maygh](http://www.ccs.neu.edu/home/amislove/publications/Maygh-EuroSys.pdf).
- Build a collaborative editor like EtherPad.
