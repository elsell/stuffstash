# Native localization evidence

[Run 36894135591](https://github.com/elsell/stuffstash/actions/runs/36894135591)
passed on iPhone 17 and iPad mini (A17 Pro) simulators at `00b58f12`.
`testLocalizedAddDraftKeepsNativeActionsAndRecovery` verifies actual React Native
RTL state, mirrored Close/Save positions, action visibility, draft entry,
in-flight disabling, rejected-save recovery with the draft preserved, and dismissal.

The retained recovery screenshots were inspected on October 1. Both show Save
on the left and Close on the right, mirrored location/detail rows, and visible
recovery content and draft. The RTL pseudolocale retains English text with
isolation markers; these images are not Arabic translation evidence. Text remains
left-aligned in parts of the form. This is acceptance of the tested native action
mirroring and recovery workflow, not a claim that every RTL layout, icon or gesture
is correct. Mixed-direction names, breadcrumbs and Map interactions remain outside
this run's coverage.

- [iPhone recovery](rtl-add-recovery-iphone.png)
- [iPad recovery](rtl-add-recovery-ipad.png)

Earlier run `36880433939` did not prove actual native RTL. This run supersedes
that limitation for the specific assertions above. Expanded-text Add/recovery
run `36876189076` at `3ba9fa2f` remains separate evidence; passing either run does
not close residual client-copy migration or physical-device verification.

## Expanded text, October 1 follow-up

[Run 36920101432](https://github.com/elsell/stuffstash/actions/runs/36920101432)
passed on iPhone 17 and iPad mini (A17 Pro), source
`673fb0b5f7bd7074edf0cfde9a98060fe3385396`, locale `en-XA`, normal text size.
The same Add/recovery fixture checks native action availability, draft entry,
in-flight disabling, failed-save recovery and dismissal. It does not exercise
production authentication or saving to a physical device.

Inspected both retained recovery images: Close and Save remain visible, expanded
labels fit their rows, the entered draft survives rejection, and the recovery
heading remains visible. The long empty-input example is ellipsized on iPhone;
this is not proof that every long title or accessible text size fits. Fixture
inventory names and the rejection diagnostic remain unchanged intentionally.

- [iPhone expanded recovery](expanded-add-recovery-iphone.png)
- [iPad expanded recovery](expanded-add-recovery-ipad.png)

This closes expanded-text Add/recovery acceptance for the tested revision. It
neither certifies the entire mobile surface nor replaces native search/approval,
VoiceOver/TalkBack, physical export, microphone or camera evidence.
