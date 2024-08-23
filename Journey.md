# Journey 

### Why did I build this
I've been studying about operating systems through Modern Operating Systems 4th edition by Andrew S. Tanenbaum, where I learned about threads and processes and Tsoding daily's video about TF-IDF, so I was curious and had to think of a project where I could utilize both concepts.

### What I've learned
I've learned alot, even after reading about processe, threads, scheduling and race conditions but to be fair I never understood that much in chapter 2 of the book in the Scheduling algorithm, only a subset of it, like have 3 differet System environments eg: Batch systems, Interactive System and Real-time Systems, I've also learned

### Challenges 
- Race Conditions: To me I think this is one of the reasons that made me almost give up, or it could just be because I have 0 idea on how to use NodeJs' Atomic Operations but I do understand Atomic operations, I've seen how easy it is to setup up mutexes in C and on Golang and it looks so much easier than having to go back and fourth between the main thread and worker thread in NodeJs and figuring out which operations to use, what index to wait for and at what index to notify the sleeping threads to begin deserializing data from a worker thread through a shared buffer to the main thread.

- Shared Array Buffers: Initially It wasn't that hard at first because I thought shared buffers would take in any word sized data until I read the docs that it needed atleast a 32-bit Array Buffer for alignment (I still don't know what that means), So I had to serialize my object data from the crawler, encode it into UTF-8, and then transform the 8-bit buffer array that holds all the data, and convert it into a 32-bit Array buffer by adding 0 padding in Little Endianess, so for the people using CPU's with a byte ordering of Big Endianess goodluck.

