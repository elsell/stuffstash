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
