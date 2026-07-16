export class BoundedTaskScheduler {
  private active = 0;
  private closed = false;
  private readonly queue: Array<{ resolve: () => void; reject: (error: Error) => void }> = [];

  constructor(private readonly limit: number) {
    if (!Number.isInteger(limit) || limit < 1) {
      throw new Error('Task scheduler limit must be a positive integer.');
    }
  }

  async schedule<T>(task: () => Promise<T>): Promise<T> {
    if (this.closed) throw new Error('Task scheduler is closed.');
    await new Promise<void>((resolve, reject) => {
      this.queue.push({ resolve, reject });
      this.drain();
    });
    try {
      return await task();
    } finally {
      this.active -= 1;
      this.drain();
    }
  }

  close(): void {
    if (this.closed) return;
    this.closed = true;
    const error = new Error('Task scheduler is closed.');
    for (const waiter of this.queue.splice(0)) waiter.reject(error);
  }

  private drain(): void {
    while (this.active < this.limit && this.queue.length > 0) {
      this.active += 1;
      this.queue.shift()?.resolve();
    }
  }
}
