# Released batch — TestFlight0.24.26 (115.1)

Frozen source2876cc19df60937ff4c0bf271f3b3a5348074f09 fixes M252 RGB conversion
rounding and M253 native disabled-state propagation. PR157 merged as
334e9ea442b6061745c6724e448021bf124e80da. Required CI35199347900 passed;
critic found no remaining production blocker. Swift conversion9472 checks,
16 focused mounted tests, TypeScript, structural and fixture installer checks
passed. Native acceptance and remaining full-suite failures are explicitly
recorded in native-full-351993.md; unrelated findings were not release gates.

Release35207859105 succeeded. iOS upload job105160238600 and changelog
job105165325139 succeeded. At2026-09-17T10:25:22Z Apple processing and exact
TestFlight0.24.26(115.1) changelog readback were verified. A Bash observer used
sleep120 between terminal-state checks; no native or release job was restarted.

TestFlight notes:
- Fixed custom color editing unexpectedly changing other RGB channels.
- Fixed locked color controls remaining enabled in the native picker.
- Known issue: the custom color picker may occasionally fail to open. Preset colors remain available.

The comprehensive142-surface/24-axis audit remains incomplete.
