# Ordinary text-entry comparison

Run [35029455242](https://github.com/elsell/stuffstash/actions/runs/35029455242),
manual `text-entry`, revision0f37869173998e361b2c2435cd47c66d0ac019b4.

The iPad mini (A17 Pro) job104584271587 completed with4/5 tests passing. The
controlled ordinary field lost characters: expected `Native draft name`, actual
`N draft name`. The inspected capture shows both the visible field and observed
draft label holding that truncated string; this is not merely a missing locator.

![Controlled input loses characters](ipad-controlled-text-loss-350294.png)

Uncontrolled ordinary, multiline, uncontrolled without accessory, and uncontrolled
without assistance all passed in this run. Earlier runs failed the uncontrolled
ordinary baseline, so one passing sample does not close it. These comparisons
also do not establish whether assistance or the accessory affects controlled input.
The next fixed diagnostic adds controlled-without-assistance and
controlled-without-accessory while preserving every existing case and exact
assertion. No production keyboard behavior is being changed as a workaround.

Phone job104584271921 remains in progress at this checkpoint. Do not infer its
outcome from the iPad results, or treat these fixtures as full native acceptance.

The expanded diagnostic passes remote TypeScript, structural and fixture-preparation
checks. Critic found no confirmed issue. Disabling assistance is a bundle comparison
(autocorrection, spelling and smart insertion), not isolation of each setting.
