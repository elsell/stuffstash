# Connected native verification — incomplete

Run [37055754739](https://github.com/elsell/stuffstash/actions/runs/37055754739)
at `f37597002cd1673488dfd4b6bc42891b48d7215d` compiled the production iPhone
client and isolated XCTest target. Real Dex, SpiceDB and API startup passed;
the unauthenticated tenant-discovery probe returned 401.

The journey entered the server address and submitted connection, then failed at
`system-auth-browser`, source line77. Neither the app web-view query nor the
SafariViewService web-view query found the expected system authentication browser.
This does not establish whether the browser failed to open or the test queried
the wrong system surface. No provider credentials were entered by this journey.
Sign-in, item persistence, relaunch and second-principal isolation remain unverified.

The retained files contain only runtime status, static stage and source line.
No authentication screenshot, response body, token or raw XCTest log is retained.

The investigation budget is exhausted for this batch. Stop hosted retries. Keep
PR #253 draft until an interactive native observation or specific new evidence
distinguishes app connection failure from system-browser discovery. The archive
request-body fix in PR #254 is independent and must ship on its own passing gates.
