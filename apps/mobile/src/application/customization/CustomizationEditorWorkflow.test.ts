import { describe, expect, it } from 'vitest';
import { CustomizationEditorWorkflow } from './CustomizationEditorWorkflow';

describe('CustomizationEditorWorkflow', () => {
  it('rejects stale loads and invalid concurrent phases', () => {
    const workflow = new CustomizationEditorWorkflow();
    const first = workflow.beginLoad();
    const second = workflow.beginLoad();
    expect(workflow.isCurrentLoad(first)).toBe(false);
    expect(workflow.isCurrentLoad(second)).toBe(true);
    expect(workflow.beginSave()).toBe(true);
    expect(workflow.beginSave()).toBe(false);
    expect(workflow.beginLifecycleConfirmation()).toBe(false);
    workflow.finishSave();
    expect(workflow.beginLifecycleConfirmation()).toBe(true);
    expect(workflow.beginLifecycleConfirmation()).toBe(false);
    expect(workflow.beginLifecycleMutation()).toBe(true);
    workflow.finishLifecycle();
    expect(workflow.beginLifecycleConfirmation()).toBe(true);
    expect(workflow.beginSave()).toBe(false);
  });

  it('authorizes and consumes a dirty exit action exactly once', () => {
    const workflow = new CustomizationEditorWorkflow();
    const action = { type: 'BACK' };
    expect(workflow.authorizeExit(action)).toBe(true);
    expect(workflow.authorizeExit(action)).toBe(false);
    expect(workflow.takeAuthorizedExit()).toBe(action);
    expect(workflow.takeAuthorizedExit()).toBeUndefined();
  });
});
