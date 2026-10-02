# Connected browser edit acceptance

Run [37063802612](https://github.com/elsell/stuffstash/actions/runs/37063802612),
connected OIDC job111026514432, passed at revision `9da6dda8` on October2,2026.
This uses the production web client with isolated real Dex, API, PostgreSQL,
SpiceDB and Garage services. It does not intercept API responses or seed a browser
session. The parent journey also completed archive round-trip and principal-isolation
checks; this evidence records only its newly added Edit observations.

At normal text size and 320×800 CSS pixels:
- Name received initial focus and keyboard input created a verified unsaved draft.
- Shift+Tab wrapped to Save and then Cancel, staying within the dialog.
- Cancel and Save fit inside the horizontal viewport; the dialog's measured
  client width and content width were both319px.
- Keyboard Enter activated Cancel. The original title remained after reload.

[Focus and geometry observations](connected-edit-keyboard.json) accompany the
[inspected screenshot](connected-edit-narrow.png). The screenshot shows the full
footer with both actions and a visible Cancel focus ring. Input text scrolls
inside its field; this sample does not certify readability of every form at320px.
No screen-reader, physical-device, or whole-app accessibility acceptance is claimed.
