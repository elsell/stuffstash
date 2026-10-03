import { assertReadActive } from '../shared/ReadRequest';
import { LabelFailure, type LabelAsset, type LabelReference, type LabelRepository } from './LabelWorkspace';

/** Resolves exclusively against the connected API. Printed hostnames never become network targets. */
export class OpenLabel {
  constructor(private readonly repository: Pick<LabelRepository, 'instance' | 'resolve'>,
    private readonly selectScope: (target: LabelAsset, signal: AbortSignal) => Promise<void>) {}
  async execute(reference: LabelReference, signal: AbortSignal): Promise<LabelAsset> {
    assertReadActive(signal);
    const instance = await this.repository.instance(signal);
    assertReadActive(signal);
    if (instance !== reference.instanceId) throw new LabelFailure('wrong_instance');
    const target = await this.repository.resolve(reference, signal);
    assertReadActive(signal);
    await this.selectScope(target, signal);
    assertReadActive(signal);
    return target;
  }
}
