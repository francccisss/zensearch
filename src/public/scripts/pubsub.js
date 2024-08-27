class PubSub {
  constructor() {
    this.events = [];
    this.consumers = [];
  }

  subscribe(event, cb) {
    if (this.events.includes(event)) {
      this.consumers.push({ event, cb });
      return;
    }
    this.events.push(event);
    this.consumers.push({ event, cb });
  }

  publish(event) {
    this.consumers.forEach((consumer, i) => {
      if (consumer.event === event) {
        consumer.cb();
      }
    });
  }

  remove_event(current_event) {
    const filtered_consumers = this.consumers.filter(
      (consumer) => consumer.event !== event,
    );

    const filtered_events = this.events.filter(
      (event) => event !== current_event,
    );
    this.consumers = filtered_consumers;
    this.events = filtered_events;
  }
}

export default PubSub;
