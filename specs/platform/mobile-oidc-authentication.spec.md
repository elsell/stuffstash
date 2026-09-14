
### Onboarding text-entry ownership

The onboarding native text field owns the text being edited. JavaScript receives
changes for validation and submission without writing each keystroke back into
the native field. This addresses the character loss reproduced by the iPhone
controlled/uncontrolled comparison in native audit run 34903318947. Seed each
field when its form step mounts; preserve the entered value through errors and
initialize the next step from the current draft. Start-over creates clean fields.
Do not apply this policy indiscriminately to search or externally controlled
editors. Keep the full-address native assertion and verify both devices again.
