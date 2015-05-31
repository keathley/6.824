<head>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
<link rel="StyleSheet" href="style.css" type="text/css">
<title>6.824 Lab 0: MapReduce</title>
</head>

## [6.824](../index.html) - Spring 2014

# 6.824 Lab 1: MapReduce

### Due: Monday Feb 10, 11:59p

* * *

### Introduction

In this lab you'll build a MapReduce library as a way to learn the Go programming language and as a way to learn about fault tolerance in distributed systems. In the first part you will write a simple MapReduce program. In the second part you will write a Master that hands out jobs to workers, and handles failures of workers. The interface to the library and the approach to fault tolerance is similar to the one described in the original [MapReduce paper](http://research.google.com/archive/mapreduce-osdi04.pdf).

### Collaboration Policy
You must write all the code you hand in for 6.824, except for code that we give you as part of the assignment. You are not allowed to look at anyone else's solution, and you are not allowed to look at code from previous years. You may discuss the assignments with other students, but you may not look at or copy each others' code. Please do not publish your code or make it available to future 6.824 students -- for example, please do not make your code visible on github.
### Software
You'll implement this lab (and all the labs) in [Go 1.2](http://www.golang.org/) (the current latest version). The Go web site contains lots of tutorial information which you may want to look at. We supply you with a non-distributed MapReduce implementation, and a partial implementation of a distributed implementation (just the boring bits).

You'll fetch the initial lab software with [git](http://git.or.cz/)(a version control system). To learn more about git, take a look at the [git user's manual](http://www.kernel.org/pub/software/scm/git/docs/user-manual.html), or, if you are already familiar with other version control systems, you may find this [CS-oriented overview of git](http://eagain.net/articles/git-for-computer-scientists/) useful.

The URL for the course git repository is<tt>git://g.csail.mit.edu/6.824-golabs-2014</tt>. To install the files in your Athena account, you need to _clone_the course repository, by running the commands below. You must use an x86 or x86\_64 Athena machine; that is, <tt>uname -a</tt> should mention <tt>i386 GNU/Linux</tt> or <tt>i686 GNU/Linux</tt> or <tt>x86_64 GNU/Linux</tt>. You can log into a public i686 Athena host with<tt>athena.dialup.mit.edu</tt>.

    $ add git
    $ git clone git://g.csail.mit.edu/6.824-golabs-2014 6.824
    $ cd 6.824
    $ ls
    src
    $

Git allows you to keep track of the changes you make to the code. For example, if you want to checkpoint your progress, you can <emph>commit</emph> your changes by running:

    $ git commit -am 'partial solution to lab 1'
    $

### Getting started

There is an input file <tt>kjv12.txt</tt> in ~/6.824/src/main, which was downloaded from [here](https://web.archive.org/web/20130530223318/http://patriot.net/~bmcgin/kjv12.txt). Compile the initial software we provide you and run it with the downloaded input file:

    $ add 6.824
    $ export GOPATH=$HOME/6.824
    $ cd ~/6.824/src/main
    $ go run wc.go master kjv12.txt sequential
    # command-line-arguments
    ./wc.go:11: missing return at end of function
    ./wc.go:15: missing return at end of function

The compiler produces two errors, because the implementation of the<tt>Map</tt> and <tt>Reduce</tt> functions is incomplete.

### Part I: Word count

Modify <tt>Map</tt> and <tt>Reduce</tt> so that <tt>wc.go</tt> reports the number of occurrences of each word in alphabetical order.

    $ go run wc.go master kjv12.txt sequential
    Split kjv12.txt
    Split read 4834757
    DoMap: read split mrtmp.kjv12.txt-0 966954
    DoMap: read split mrtmp.kjv12.txt-1 966953
    DoMap: read split mrtmp.kjv12.txt-2 966951
    DoMap: read split mrtmp.kjv12.txt-3 966955
    DoMap: read split mrtmp.kjv12.txt-4 966944
    DoReduce: read mrtmp.kjv12.txt-0-0
    DoReduce: read mrtmp.kjv12.txt-1-0
    DoReduce: read mrtmp.kjv12.txt-2-0
    DoReduce: read mrtmp.kjv12.txt-3-0
    DoReduce: read mrtmp.kjv12.txt-4-0
    DoReduce: read mrtmp.kjv12.txt-0-1
    DoReduce: read mrtmp.kjv12.txt-1-1
    DoReduce: read mrtmp.kjv12.txt-2-1
    DoReduce: read mrtmp.kjv12.txt-3-1
    DoReduce: read mrtmp.kjv12.txt-4-1
    DoReduce: read mrtmp.kjv12.txt-0-2
    DoReduce: read mrtmp.kjv12.txt-1-2
    DoReduce: read mrtmp.kjv12.txt-2-2
    DoReduce: read mrtmp.kjv12.txt-3-2
    DoReduce: read mrtmp.kjv12.txt-4-2
    Merge phaseMerge: read mrtmp.kjv12.txt-res-0
    Merge: read mrtmp.kjv12.txt-res-1
    Merge: read mrtmp.kjv12.txt-res-2

The output will be in the file "mrtmp.kjv12.txt". Your implementation is correct if the following command produces the following top 10 words:

    $ sort -n -k2 mrtmp.kjv12.txt | tail -10
    unto: 8940
    he: 9666
    shall: 9760
    in: 12334
    that: 12577
    And: 12846
    to: 13384
    of: 34434
    and: 38850
    the: 62075

To make testing easy for you, run:

    $ ./test-wc.sh

and it will report if your solution is correct or not.

Before you start coding read Section 2 of the [MapReduce paper](http://research.google.com/archive/mapreduce-osdi04.pdf) and our code for mapreduce, which is in <tt>mapreduce.go</tt> in package <tt>mapreduce</tt>. In particular, you want to read the code of the function <tt>RunSingle</tt> and the functions it calls. This well help you to understand what MapReduce does and to learn Go by example.

Once you understand this code, implement <tt>Map</tt> and <tt>Reduce</tt> in<tt>wc.go</tt>.

Hint: you can use [<tt>strings.FieldsFunc</tt>](http://golang.org/pkg/strings/#FieldsFunc)to split a string into components.

Hint: for the purposes of this exercise, you can consider a word to be any contiguous sequence of letters, as determined by [<tt>unicode.IsLetter</tt>](http://golang.org/pkg/unicode/#IsLetter). A good read on what strings are in Go is the [Go Blog on strings](http://blog.golang.org/strings).

Hint: the strconv package (http://golang.org/pkg/strconv/) is handy to convert strings to integers etc.

You can remove the output file and all intermediate files with:

    $ rm mrtmp.*

### Part II: Distributing MapReduce jobs

In this part you will design and implement a master who distributes jobs to a set of workers. We give you the code for the RPC messages (see <tt>common.go</tt> in the <tt>mapreduce</tt> package) and the code for a worker (see <tt>worker.go</tt> in the <tt>mapreduce</tt> package).

Your job is to complete <tt>master.go</tt> in the <tt>mapreduce</tt>package. In particular, the <tt>RunMaster()</tt> function in<tt>master.go</tt> should return only when all of the map and reduce tasks have been executed. This function will be invoked from the <tt>Run()</tt>function in <tt>mapreduce.go</tt>.

The code in <tt>mapreduce.go</tt> already implements the<tt>MapReduce.Register</tt> RPC function for you, and passes the new worker's information to <tt>mr.registerChannel</tt>. You should process new worker registrations by reading from this channel.

Information about the MapReduce job is in the <tt>MapReduce</tt> struct, defined in <tt>mapreduce.go</tt>. Modify the <tt>MapReduce</tt> struct to keep track of any additional state (e.g., the set of available workers), and initialize this additional state in the <tt>InitMapReduce()</tt>function. The master does not need to know which Map or Reduce functions are being used for the job; the workers will take care of executing the right code for Map or Reduce.

In Part II, you don't have worry about failures of workers. You are done with Part II when your implementation passes the first test set in<tt>test_test.go</tt> in the <tt>mapreduce</tt> package.

<tt>test_test.go</tt> uses Go's unit testing. From now on all exercises (including subsequent labs) will use it, but you can always run the actual programs from the <tt>main</tt> directory. You run unit tests in a package directory as follows:

    $ go test

The master should send RPCs to the workers in parallel so that the workers can work on jobs concurrently. You will find the <tt>go</tt> statement useful for this purpose and the [Go RPC documentation](http://golang.org/pkg/net/rpc/).

The master may have to wait for a worker to finish before it can hand out more jobs. You may find channels useful to synchronize threads that are waiting for reply with the master once the reply arrives. Channels are explained in the document on [Concurrency in Go](http://golang.org/doc/effective_go.html#concurrency).

We've given you code that sends RPCs via "UNIX-domain sockets". This means that RPCs only work between processes on the same machine. It would be easy to convert the code to use TCP/IP-based RPC instead, so that it would communicate between machines; you'd have to change the first argument to calls to Listen() and Dial() to "tcp" instead of "unix", and the second argument to a port number like ":5100". You will need a shared distributed file system.

The easiest way to track down bugs is to insert log.Printf() statements, collect the output in a file with <tt>go test &gt;
out</tt>, and then think about whether the output matches your understanding of how your code should behave. The last step is the most important.

### Part III: Handling worker failures

In this part you will make the master handle workers failures. computing a job. In MapReduce handling failures of workers is relatively straightforward, because the workers don't have persistent state. If the worker fails, any RPCs that the master issued to that worker will fail (e.g., due to a timeout). Thus, if the master's RPC to the worker fails, the master should re-assign the job given to the failed worker to another worker.

An RPC failure doesn't necessarily mean that the worker failed; the worker may just be unreachable but still computing. Thus, it may happen that two workers receive the same job and compute it. However, because jobs are idempotent, it doesn't matter if the same job is computed twice---both times it will generate the same output. So, you don't have to anything special for this case. (Our tests never fail workers in the middle of job, so you don't even have to worry about several workers writing to the same output file.)

You don't have to handle failures of the master; we will assume it won't fail. Making the master fault-tolerant is more difficult because it keeps persistent state that must be replicated to make the master fault tolerant. Keeping replicated state consistent in the presence of failures is challenging. Much of the later labs are devoted to this challenge.

Your implementation must pass the two remaining test cases in<tt>test_test.go</tt>. The first case tests the failure of one worker. The second test case tests handling of many failures of workers. Periodically, the test cases starts new workers that the master can use to make forward progress, but these workers fail after handling a few jobs.

### Handin procedure

Submit your code via the class's submission website, located here:

[https://ydmao.scripts.mit.edu:444/6.824/handin.py](https://ydmao.scripts.mit.edu:444/6.824/handin.py)

You may use your MIT Certificate or request an API key via email to log in for the first time. Your API key (XXX) is displayed once you logged in, which can be used to upload lab1 from the console as follows.

    $ cd ~/6.824
    $ echo XXX > api.key
    $ make lab1

You can check the submission website to check if your submission is successful.

You will receive full credit if your software passes the <tt>test_test.go</tt> tests when we run your software on our machines. We will use the timestamp of your **last** submission for the purpose of calculating late days.

* * *
<address>
Please post questions on <a href="http://piazza.com">Piazza</a>.
<p>

</p>
</address>
