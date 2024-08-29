async function polling(poll: any, time: number) {
  while (poll) {
    await new Promise((resolved) => {
      setTimeout(() => {}, time * 1000);
    });
  }
}
