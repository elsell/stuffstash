import { expect, it } from 'vitest';
import { expirationDateLabel } from './expirationPresentation';
const context={state:'upcoming',trackingEnabled:true,advanceDays:30,timezone:'America/Los_Angeles'} as const;
it('says today in the saved timezone through the final local day',()=>{
 expect(expirationDateLabel({date:'2026-09-11',precision:'day'},context,new Date('2026-09-12T02:00:00Z'),'en-US')).toContain('Expires today');
});
it('preserves month-only precision when its final day is today',()=>{
 const label=expirationDateLabel({date:'2026-09',precision:'month'},context,new Date('2026-10-01T02:00:00Z'),'en-US');
 expect(label).toContain('Expires today');expect(label).toContain('September 2026');expect(label).not.toContain('September 30');
});
it('retained dates explicitly show disabled tracking',()=>{expect(expirationDateLabel({date:'2026-09',precision:'month'},{...context,trackingEnabled:false},new Date('2026-10-01T02:00:00Z'),'en-US')).toContain('tracking disabled');});
