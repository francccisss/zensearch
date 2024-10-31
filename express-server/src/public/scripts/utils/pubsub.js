class PubSub {
  constructor() {
    this.events = {};
  }

  subscribe(event, ...args) {
    try {
      args.forEach((cb) => {
        if (typeof cb !== "function") {
          throw new Error("Expecting a callback function.");
        }
      });
      if (!this.events[event]) {
        this.events[event] = [];
      }
      this.events[event] = [...this.events[event], ...args];
    } catch (err) {
      console.error(err.message);
    }
  }

  // TODO handle asynchronous calls lad
  publish(event, updateData) {
    if (!this.events[event]) {
      console.error("Event does not exist: %s", event);
      return;
    }
    this.events[event].forEach((cb) => {
      cb(updateData);
    });
  }
}

const pubsub = new PubSub();

export default pubsub;
