# NFCX

<p align="center"><img src="docs/images/nfcx-logo.png" alt="NFCX logo" width="144"></p>

<p align="center"><strong>A focused, cross-platform desktop workbench for authorized NFC card work.</strong></p>

<p align="center"><a href="README.md">English</a> | <a href="README.zh-CN.md">简体中文</a></p>

NFCX brings reader discovery, card information, reads, protected writes, raw dumps, key management, and recovery workflows into one desktop application for macOS, Windows, and Linux. Use it only with MIFARE Classic cards you own or are authorized to test.

Website: [nfcx.tools](https://nfcx.tools) · Downloads: [GitHub Releases](https://github.com/BennyThink/NFCX/releases)

## What NFCX can do

- Discover and connect NFC readers; PN532 UART can also be selected with an exact connection string.
- Detect cards and show UID, ATQA, SAK, and a conservative card-type indication.
- Work with MIFARE Classic 1K: authenticate with Key A or Key B, read blocks, edit data, and write changes.
- Save and load compatible raw dumps (`.bin` / `.mfd`) with sidecar metadata; restore with capacity, BCC, access-bit, and read-back checks.
- Scan common keys and manage a local key catalog.
- Run the integrated recovery sequence on the selected PN532 UART profile: common keys, Darkside, Nested, Hardnested, and read verification.
- Use the guarded 4-byte UID/block 0 workflow for supported CUID/Gen2 and Gen1A cards. It is off by default and never targets unknown card types blindly.

## Supported readers

| Reader | Connection / backend | Status | Platforms |
| --- | --- | --- | --- |
| PN532 + FT232RL | Serial, `pn532_uart` via libnfc | **Tested** for discovery, Classic read/write, dump/restore, and recovery | macOS, Windows, Linux builds |
| Other PN532 UART adapters | Serial, `pn532_uart` via libnfc | **Expected compatible**; not hardware-tested by this project | Depends on adapter driver and serial permissions |
| ACR122U | USB / PC/SC or libnfc | **Unsupported in this release**: no NFCX hardware validation or enabled release profile | — |
| ACR1552U | USB / vendor PC/SC driver | **Unsupported in this release**: no NFCX hardware validation or enabled release profile | — |
| Other libnfc devices | Varies | **Unsupported** until an explicit NFCX validation profile exists | — |

“Expected compatible” means libnfc may recognize the hardware; it does not mean every NFCX workflow has been verified. Key recovery and special-UID workflows are restricted to the validated PN532 UART profile.

## Drivers and requirements

NFCX packages its NFC runtime. You do not need to install libnfc, mfoc, mfcuk, or a separate command-line NFC stack.

- **PN532 + FT232RL:** install the FTDI virtual-COM/serial driver if your OS does not expose the adapter. On Linux, your account generally needs read/write serial access (commonly the `dialout` group).
- **macOS:** an unsigned release can show a Gatekeeper warning. Signed and notarized releases avoid this once release credentials are configured.
- **Windows:** install the adapter’s serial driver before starting NFCX. Do not replace the device with a generic USB driver for the PN532 UART path.
- **ACR readers:** they are not supported by this release; do not assume a system PC/SC driver is sufficient.

## Download and install

1. Open [GitHub Releases](https://github.com/BennyThink/NFCX/releases).
2. Download the archive or installer for your operating system.
3. Install or unpack it, then install the reader driver if needed.
4. Connect the reader and start NFCX.

Release targets are macOS arm64, Windows amd64, and Linux amd64. Linux packages include AppImage and tar.gz options; system GTK3 and WebKitGTK 4.1 are still required for the desktop UI.

## Quick start

1. Connect your NFC reader and launch NFCX.
2. Refresh the reader list or enter the PN532 UART connection string.
3. Put an authorized card on the reader and select **Scan Card**.
4. Review the card information, then use **Read**, **Dump**, **Write**, or **Key Recovery** as appropriate.

Writing is deliberately guarded. NFCX validates dump geometry, BCC and access bits, authenticates before writing, and reads every written block back. Block 0 writes require an explicit special-card workflow.

## Screenshots

Maintainers will add real screenshots to [`docs/images/`](docs/images/): `main-window.png`, `card-scan.png`, `mifare-tools.png`, and `about.png`. Until then, NFCX does not substitute fictional UI artwork.

## Privacy and anonymous telemetry

Anonymous telemetry is optional. You choose at first launch and can change the choice at any time in **About**. When enabled, NFCX sends only an anonymous installation identifier, NFCX version, coarse operating-system type, allowlisted feature events, and event time to understand usage and platform distribution.

NFCX does **not** collect card UIDs, Key A/Key B values, dumps, card contents, reader identifiers, usernames, device names, machine IDs, file paths, logs, IP addresses, or other personal information. The collector does not retain raw request headers or IP data. See the [telemetry specification](docs/specs/14-telemetry.md) for implementation details.

## License and third-party software

NFCX source code is released under the [MIT License](LICENSE). NFCX dynamically links LGPL-3.0-or-later libnfc and redistributes separate GPL-2.0-or-later recovery executables (mfoc, mfcuk, and mfoc-hardnested), plus the BSD-2-Clause `nfc-mfsetuid` utility. These remain independently licensed programs; their notices, source locations, pinned versions, and NFCX patches are included in release packages.

Read [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) before redistributing NFCX or its runtime.

## Project links

- [Website](https://nfcx.tools)
- [GitHub repository](https://github.com/BennyThink/NFCX)
- [Releases](https://github.com/BennyThink/NFCX/releases)
- [Documentation index](docs/README.md)

Use NFCX only with cards and systems you own or are explicitly authorized to test.
