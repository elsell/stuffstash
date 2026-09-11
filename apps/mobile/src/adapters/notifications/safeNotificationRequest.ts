import { assertReadActive } from '../../application/shared/ReadRequest';
import { NotificationFailure } from '../../application/notifications/NotificationFailure';
import { StuffStashAPIError } from '@stuff-stash/api-client';

export async function safeNotificationRequest<T>(operation: () => Promise<T>, signal?: AbortSignal): Promise<T> {
    assertReadActive(signal);
    try { const result = await operation(); assertReadActive(signal); return result; }
    catch (error) {
      assertReadActive(signal);
      if (error instanceof Error && error.name === 'AbortError') throw error;
      throw new NotificationFailure(error instanceof StuffStashAPIError
        ? error.status === 401 ? 'authentication-required' : error.status === 403 ? 'permission-denied'
        : error.status === 404 ? 'not-found' : error.status === 409 ? 'conflict'
        : error.status === 400 || error.status === 422 ? 'invalid' : 'unavailable'
        : 'unavailable');
    }
  }
