import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { PrintCopiesControl } from './PrintCopiesControl.ios';
it('accepts positive safe counts but rejects locked, invalid and retired native events', async () => {
  const h = new MobileRenderHarness(); const values: string[] = [];
  try {
    await h.render(<PrintCopiesControl label="Copies" value="1" disabled={false} onChange={value => values.push(value)} />);
    const stepper = h.byType('SwiftUIStepper')!;
    const change = stepper.props.onValueChange;
    await h.run(() => { change(2); change(0); change(1.5); change(Number.MAX_SAFE_INTEGER + 1); });
    expect(values).toEqual(['2']);
    await h.render(<PrintCopiesControl label="Copies" value="2" disabled onChange={value => values.push(value)} />);
    await h.run(() => change(3)); expect(values).toEqual(['2']);
    await h.render(<></>); await h.run(() => change(4)); expect(values).toEqual(['2']);
  } finally { await h.unmount(); }
});
