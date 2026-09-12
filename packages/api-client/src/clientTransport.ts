import type { components } from './generated/schema';
type Result<T> = { data?: T; error?: components['schemas']['ErrorEnvelope']; response: Response };
export interface ClientTransport {
 headers(): Promise<Record<string, string>>;
 unwrap<T>(request: Promise<Result<T>>): Promise<T>;
}
