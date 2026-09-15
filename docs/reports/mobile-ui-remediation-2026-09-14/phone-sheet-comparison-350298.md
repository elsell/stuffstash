# iPhone sheet structure comparison

Inspected artifact10423121514 from run35029854251, revision
e8b3d42dccf3f13428fb26bbb1cfd85ea8b0dd9e. These are normal-size diagnostic
fixtures, not screenshots of the production filter screen.

The nested View/ScrollView plus footer variant shows a blank content area while
its footer remains visible: [capture](phone-nested-footer-blank-350298.png).
The direct ScrollView plus overlay footer variant shows all four rows and both
commands: [capture](phone-direct-footer-350298.png). The former test fails because
Diagnostic Tags is absent, not because its footer label is wrong.

Production NativeFilterSheet uses the direct-scroll structure and reserves a
measured, opaque footer area. The comparison reinforces that structural choice;
it is not sufficient evidence of a new production filter regression. Production
tag-list, search/keyboard and date-range journeys remain their own acceptance gates.
Do not turn the diagnostic's expected comparison result into a claim that every
production sheet is broken, or treat its passing counterpart as whole-app approval.
