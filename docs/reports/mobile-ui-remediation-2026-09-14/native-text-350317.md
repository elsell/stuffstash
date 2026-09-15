# Expanded text-entry diagnostic — partial results

Run35031744887 tests source81e91f749268770ab984b80d49009bb9e7cebace.
iPhone job104594897722 completed with **4/7 passing** on September15,2026.
iPad remains running. Evidence: terminal XCTest log; screenshots not inspected.

| Scenario | iPhone outcome |
| --- | --- |
| Controlled without accessory | Failed: `Ne draft name` |
| Controlled without assistance | Passed |
| Ordinary controlled | Failed: `Ndraft name` |
| Ordinary multiline | Passed |
| Ordinary uncontrolled single line | Failed: `Nraft name` |
| Uncontrolled without accessory | Passed |
| Uncontrolled without assistance | Passed |

All expected `Native draft name`. Disabling the accessory does not reliably remove
the failure, and controlled state is not a sufficient explanation because the
uncontrolled baseline also fails. Both no-assistance cases passed in this sample;
that variant disables autocorrection, spell checking and smart insert together.
This narrows further investigation but does not establish causation or justify
removing those features across production inputs. Earlier runs varied. No text
entry finding is closed, and no production keyboard change follows this sample.
