# Pinned MFOC source

NFCX builds the upstream `nfc-tools/mfoc` `mfoc-0.10.7` tag at commit
`290a0759567d4df4840d9133d663ac88246b373a`. The release source archive,
SHA-256, license and required libnfc version are recorded in `mfoc.lock`.

MFOC is GPL-2.0-or-later. Runtime packaging must include its `COPYING` file,
the applied NFCX patches, and provide the corresponding source through the
recorded archive URL. NFCX renames the `usage` function's `errno` parameter to
avoid a modern MinGW-w64 macro collision. It builds MFOC against the same
pinned, patched libnfc 1.8.0 SDK used by the in-process CGO binding.

The second patch, `0002-add-high-performance-and-x64-modes.patch`, adds the two
NFCX runtime modes documented in `docs/notes/26-mfoc-performance-modes.md`:

- `-H` high-performance mode scans the Nested probe window on every CPU core
  (`NFCX_MFOC_THREADS` overrides the worker count) and merges the per-worker
  candidate keys. The recovered candidate set is identical to the serial scan.
- `-X` x64 mode evaluates the three nonce parity bits inside one 64-bit
  expression using `__builtin_parity()`; 32-bit builds reject the flag.

The patch series is applied in the order listed by `MFOC_PATCH_SERIES`, and
`make check` builds `tests/nfcx-perf-test.c`, which cross-checks the wide parity
path against upstream `valid_nonce()`, verifies that the parallel scanner
recovers the same candidates as the serial one, and reports the speed-up.
