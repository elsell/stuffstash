# Expanded text-entry diagnostic — terminal results

Run35031744887 tests source81e91f749268770ab984b80d49009bb9e7cebace.
iPhone job104594897722 completed with **4/7 passing** on September15,2026.
iPad job104594897956 completed with **3/7 passing**. Evidence: terminal XCTest
logs; screenshots not inspected.

| Scenario | iPhone outcome |
| --- | --- |
| Controlled without accessory | Failed: `Ne draft name` |
| Controlled without assistance | Passed |
| Ordinary controlled | Failed: `Ndraft name` |
| Ordinary multiline | Passed |
| Ordinary uncontrolled single line | Failed: `Nraft name` |
| Uncontrolled without accessory | Passed |
| Uncontrolled without assistance | Passed |

iPad ordinary controlled and uncontrolled single-line inputs both produced
`N draft name`; multiline, uncontrolled without accessory, and uncontrolled
without assistance passed. Controlled-without-accessory could not launch through
Xcode, and controlled-without-assistance failed to acquire an app background
assertion. Those two are infrastructure/readiness failures, not input observations.
Do not count them as evidence that those control configurations lost text.

All expected `Native draft name`. Disabling the accessory does not reliably remove
the failure, and controlled state is not a sufficient explanation because the
uncontrolled baseline also fails. Both no-assistance cases passed in this sample;
that variant disables autocorrection, spell checking and smart insert together.
This narrows further investigation but does not establish causation or justify
removing those features across production inputs. Earlier runs varied. No text
entry finding is closed, and no production keyboard change follows this sample.
