import type { ConsumeMessage } from "amqplib";

// const SIZE = 450;

// end is write
// start is read
// if queue is full, end === queue length where every index is in-use
// if queue is empty start === end
// every read set previously read index as null, so when write comes across
// a null value, it can write into it

class CircularBuffer {
  private N: number = 0;
  queue: Array<ConsumeMessage | null>;
  writeIndex: number = 0;
  readIndex: number = 0;
  constructor(size: number) {
    this.N = size;
    this.queue = new Array(size).fill(null);
    this.readIndex = 0;
    this.writeIndex = 0;
  }

  read(): ConsumeMessage | null {
    if (this.readIndex === this.writeIndex) {
      console.log("Buffer is empty");
    }
    const data = this.queue[this.readIndex];
    this.queue[this.readIndex] = null;
    this.readIndex = (this.readIndex + 1) % this.N;
    return data;
  }

  write(data: ConsumeMessage) {
    if (this.writeIndex === this.queue.length - 1) {
      console.log("Write index reached the end: WRAP");
    }
    if (this.queue[this.writeIndex] !== null) {
      console.log("Buffer Overflow: Either overwrite or pause consumer");
    }
    this.queue[this.writeIndex] = data;
    this.writeIndex = (this.writeIndex + 1) % this.N;
  }
  inUseSize(): number {
    let nonNullCounter = 0;
    for (let i = 0; i < this.queue.length; i++) {
      if (this.queue[i] !== null) nonNullCounter++;
    }
    return nonNullCounter;
  }
}

//const CircBuffer = new CircularBuffer(SIZE);

export default CircularBuffer;
