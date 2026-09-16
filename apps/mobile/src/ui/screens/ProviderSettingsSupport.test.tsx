import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { ProviderStateView } from './ProviderSettingsSupport';

it.each(['voice setup', 'listen settings', 'provider profiles', 'provider profile', 'credential settings', 'prompt guidance'])('identifies %s during loading and failure and retains retry', async taskLabel => {
  const h = new MobileRenderHarness(); let retries = 0;
  try {
    await h.render(<ProviderStateView taskLabel={taskLabel} state={{ status: 'loading' }} onRetry={async () => { retries++; }} />);
    expect(h.allText()).toContain(`Loading ${taskLabel}`);
    expect(h.all().some(node => node.props.accessibilityRole === 'progressbar')).toBe(true);
    await h.render(<ProviderStateView taskLabel={taskLabel} state={{ status: 'error', message: 'Connection unavailable' }} onRetry={async () => { retries++; }} />);
    expect(h.allText()).toContain(`Could not load ${taskLabel}`);
    expect(h.allText()).toContain('Connection unavailable');
    await h.press(h.byLabel('Retry')); expect(retries).toBe(1);
  } finally { await h.unmount(); }
});
