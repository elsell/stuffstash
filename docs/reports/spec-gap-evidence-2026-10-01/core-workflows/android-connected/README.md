# Connected Android core walkthrough — October 3, 2026

Normal application APK from [build 37110497691](https://github.com/elsell/stuffstash/actions/runs/37110497691), source `a93b524a387e8d01ad28cf58b34002cdb880e5d6`. The installed APK SHA-256 was rechecked as `650307b6db7fa59c33c2b30723e6e3899b15c5a5b0946e986247940916275e20`.

Android 16/API 36 emulator, 1080×2400 pixels, density 420, font scale 1.0, light appearance. The app used its existing authenticated session against the isolated Paul Dex/API/PostgreSQL/SpiceDB/Garage test server. No API interception or injected in-memory app data was used. The exact test CA and hostname mapping were restored to the emulator; TLS verification remained enabled. This was not another sign-in/isolation acceptance run.

## Three scoped verification gaps

1. **Connected Edit and persistence:** Browse searched for `flashlight`; Details opened the synthetic Connected audit flashlight. Edit accepted the description suffix ` Android core audit.` and Save returned to Details with that text. After force-stop and relaunch, opening the item from Home still showed the saved description. [Draft screenshot](edit-draft.png), [reopened Details observation](persisted-details.txt).
2. **Connected Move selection, cancellation and commit:** choosing Inventory root updated the proposal; Cancel returned to Details with Audit garage unchanged. Reopening Move showed Audit garage as the current location. Selecting root and pressing Move changed Details to No location. Android Back returned to the retained flashlight search with the changed item. After app relaunch the root location persisted. A subsequent explicit Move, using a garage search, restored Audit garage. [Selected proposal](move-selected.png), [committed move](moved.png), [search return](returned-search.txt), [restored location](restored-location.txt).
3. **Connected Filters and navigation return:** the in-place Type menu selected Places; Show results returned one place and an applied-filter indicator. Opening Audit garage and pressing Android Back retained that filter and result. Removing the filter returned six results. [Filter draft](filter-draft.png), [filtered return](filter-return.txt), [reset](reset-filter.txt).

The item retains the synthetic description suffix as audit test data. Its location was restored and the Browse filter was cleared. No user inventory was changed. The owned emulator was stopped after the walkthrough.

## Visual review and limits

Inspected full-screen Browse, Map, Edit, Move and Filters captures. Edit and Move expose separate header commit/cancel actions; their content and commands fit this normal-text viewport. Filters shows the primary completion action and secondary cancel clear of system navigation. These observations do not certify every state or establish iOS quality.

The List/Map selector shifts horizontally between [List](browse.png) and [Map](map.png), alongside changes in available toolbar actions. This is runtime-observed; user confirmation was requested before treating it as a new fix. It remains open, not a passing stability claim.

The earlier intermittent Home/Browse refresh failure did not occur during this bounded walkthrough. That does not diagnose or close it. No new VoiceOver/TalkBack, enlarged-text, RTL, physical device, camera/microphone, rejected mutation, destination creation or timing acceptance is claimed. These three workflow samples supplement the broader audit; they do not complete it.
