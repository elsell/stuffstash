# Safe flag diagnostics

The CLI reports the registered option responsible for a flag conversion failure,
the expected value type and an action to correct it. Messages never include the
supplied value or raw flag-library error. They retain the usage category and exit
status 2. Execution and local help share the same diagnostic parser.

The parser wraps the actual registered flag values, forwarding their behavior and
recording a failed Set directly. It does not parse error strings, reimplement
numeric conversions, or maintain a second flag catalog. Type guidance comes from
registered flag.Getter values; custom validators receive safe generic value
guidance. Boolean arity, explicit false values, FlagSet.Visit, repeated options
(last value wins), signed numeric/float parsing and positional partitioning remain
unchanged. Internal flag diagnostics stay suppressed.

Range validation distinguishes copies below 1 from template versions above the
unsigned 32-bit maximum. Copies require an integer of at least 1. Template versions
accept 0 through 4294967295, where 0 retains the existing default-selection behavior.
These messages identify the option and bound without echoing its value. No range
or operation behavior changes in this slice.

Critical tests cover integer, unsigned, float and boolean conversion failures with
secret-like inputs, help-path failures, independent range messages, normal mixed
options, repeated last-value selection, explicit false and scope flag visitation.
Bootstrap tests retain usage exit 2 and check diagnostics do not expose values.
