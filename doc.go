/*
Package worker implements the core of Worker.

The main coordinating component in the core is the Processor, and is a good
place to start when exploring the code. It gets the jobs off the job queue and
calls the right things in order to run it. The ProcessorPool starts up the
required number of Processors to get the concurrency that's wanted for a single
Worker instance.
*/
package worker
