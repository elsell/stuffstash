import { createAuthenticatedTransport } from '../authenticatedTransport';
import type { StuffStashClientOptions } from '../stuffStashClient';
import { PerformanceReporter, type PerformanceReporterOptions } from './performanceReporter';

export interface ApiPerformanceReporterOptions extends Omit<PerformanceReporterOptions, 'send'> {
  connection: StuffStashClientOptions;
}

/** Use an unobserved transport to prevent recursive reporting of telemetry delivery. */
export function createApiPerformanceReporter(options: ApiPerformanceReporterOptions): PerformanceReporter {
  const transport = createAuthenticatedTransport(options.connection);
  return new PerformanceReporter({
    clock: options.clock,
    scheduler: options.scheduler,
    capacity: options.capacity,
    batchSize: options.batchSize,
    send: async (batch, signal) => {
      const { data, error, response } = await transport.POST('/client-telemetry', {
        body: { measurements: [...batch] }, signal
      });
      if (error || !response.ok || data?.data.accepted !== batch.length) {
        throw new Error('Performance delivery failed');
      }
    }
  });
}
