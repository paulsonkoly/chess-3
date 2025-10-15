# Distributed tuner design

## Motivation

The existing tuner albeit multi-threaded runs on a single machine and loads all the EPD data to memory. This was a convenient approach but until now, but as the EPD filesize and the number of coeffs grew the tuner became increasingly slow and on certain machines it would fail due to running out of memory.

## Current system

The current system works on a single machine, and implements the ADAM algorithm on multiple threads. The entire EPD data is loaded memory represented as a go slice. An entry consists of an EPD position string and a WDL label. The algorithm iterates the data, one iteration is called an epoch and after each epoch it re-shuffles the data-set.
Within an epoch we split the work into batches. The optimiser runs on a single batch with the current values of the coefficients, and produces the batch gradient as a result. We add the batch gradients to the coeff set on the completion of the batch.
A batch is split into chunks based on the number of threads. Chunks are dispatched to workers, that produce the sub-result gradients. These are then summed when all threads finished with their corresponding chunk, to produce the batch gradient.

```go
epd := ReadEPDFile("file.epd")

coeffs := intialCeoffs{}

for epoch := 1; true; epoch ++ {
  for batch := range epd.Batches() {
    wg := sync.WaitGroup{}
    threadGrads := make([]Coeffs, numThreads) 

    for i, chunk := batch.Split(numThreads) {
       wg.Add(1)
       go func() {
          threadGrads[i] = calcGradient(chunk, coeffs)
          wg.Done()
       }
    }

    wg.Wait()
    batchGrad := Coeffs{}
    for _, grad := range threadGrads {
        batchGrad.Add(grad)
    }

    coeffs.addADAM(batchGrad, momentum, velocity, lr ...)
  }
  epd = epd.shuffle()
}
```

## Proposed design

### Goals

The new system should adhere to the followings:

- can utilise computing resources of multiple machines, thus runs faster than the current implementation
- on a single machine would not cause out of memory errors
- does not degrade the optimisation algorithm convergence
- resilient to network / client machine failures

### Plan

A mostly stateless client server architecture based on the simplest design that can fulfill our goals. There would be one server - the orchestrator, and multiple clients - the workers. The server does not register connecting clients, or keep client state. This simplifies the architecture and makes it resilient to arbitrary client failures. The server has to however keep track of jobs needed to be done. The main idea is the following protocol between clients and server:

```
client --- request job ---> server
client --- job results ---> server
```

The server never initiates communication with the client. It simply distributes jobs to be done, and registers job results when a client finishes one. In case of a client failure the server would have a time out detection on the job, and simply would give it out to the next client upon request.

Jobs map to chunks in our original design.
Once a server registers that a batch is fully completed, it moves to the next batch and updates the current coefficients, in line with the original implementation.
The server cannot give jobs to requesting clients if there are no more chunks left in a batch to give out, and thus has to stall clients when this happens. This is needed because new jobs would require the updated coefficients.

We can implement the Job as

```go
type Job struct {
  jobUUID string // a unique JOB id, different if the same job is given to other clients or the same client again
  epoch int // identifies the shuffle sequence of the EPD file
  checksum string // the checksum of the shuffled chunk
  coefficients []float64 // the current coefficients
  start int // line index of the chunk start
  end int // line index of the chunk end
}

```

End the Result structure would look like

```go
type Result struct {
   jobUUID string // same UUID that the client received
   checksum // the chunk checksum
   gradients []float64 // the computed gradients
}
```



#### EPD shuffle

The EPD files are distributed to new joining clients that can request a streaming API to obtain their local copy of the EPD data. They can also cache the EPD on local hardrive and verify their content on startup with checksum, avoiding re-downloads on client startup.

At the end of the epoch there is a random shuffle of the EPD data. The shuffle needs to be consistent across all clients. The current epoch number uniquely identifies the shuffle order. The epoch number has to be part of the job structure thus the job incorporates the EPD order.

Once the EPD is shuffled, the client can store a local copy of the shuffled data to a new epd file..

Once the shuffled data is stored on disk, the client can unload it from memory and on Job request it can reload the relevant part to memory.

The shuffling can be based on lines, instead of parsed EPD entries thus we should get out of memory errors when a client does the shuffling. The current design keeps board representation in the slice that holds the EPD data, because it loads only once.

The shuffle is verified by the Job checksum, which is sent back to the server on completion. Then the server can also verify the result.
