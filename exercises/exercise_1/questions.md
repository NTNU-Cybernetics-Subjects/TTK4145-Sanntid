Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> 
Concurrency is when two or more processes shares running time on the CPU, it will look like they run simultaneously, but in reality the scheduler will perform context switching on the two processes in a kind of multitasking manner. 

In parallelism the processes literally run simultaneously (parallel). This can be achieved by a multicore processor which can run processes parallel.

What is the difference between a *race condition* and a *data race*? 
> 
Data race is when two or more threads access a shared variable at the same time. For example one thread reads a variable while another one writes to it. Race condition is about which thread manipulates the variable first. 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> 

A scheduler changes which program, or task that gets running time one the processor. The schedulers task is to multitask between the different tasks that needs to be executed such that "more than one thing can happen at once". The operation of changing the running task on the cpu is called context switching, and roughly it works by storing the progress of the currently running task on the heap (or stack?) and then loads the new task into the CPU. When the first task gets running time again it loads in the stored progress and continue executing the task.  


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> 
Multiple threads helps keep the code untangled. It gives the programmer the option to separate code, or jobs, with different objectives into different places that makes the code easier to read, understand and implement. 

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> 
Coroutines is when the program executes another job(task) when it encounters a wait statement instead of just waiting for it to finish. For example while you are waiting for some network communication, you sum up the first 100000 Fibonacci numbers. When you receive the network request you were waiting for you stop summing the Fibonacci numbers, handles the request then you go back to Fibonacci summation. 

This is different from multithreading because you utilize the waiting time for some task to finish or happen to do another task, instead of context switching different jobs so that they run "simultaneously".

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> 
Concurrent programming makes the development easier in the sense that the code gets more transparent with more logical architecture. It also makes it harder because because you have to syncronize the threads, take race condition into consideration etc. It also makes the code much harder to debug. 

What do you think is best - *shared variables* or *message passing*?
> 
It depends on the purpose, shared variables will be faster because all parts involved reads and writes from/to the same memory. It is however more prone to problems, because you need to handle access to the shared resource to not get unpredictable results. Message passing will keep everything separated which makes the program more intuitive, but as mentioned is slower than shared variables. 



