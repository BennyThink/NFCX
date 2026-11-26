# MFOC 高性能与 x64 模式

日期：2026-11-22

## 背景

Nested 攻击的耗时几乎全部落在 `lfsr_recovery32()`：对 `median ± tolerance` 窗口内每一个通过 3 bit 奇偶校验的 nonce 各调用一次，每次都要重新分配 8 MiB 临时表并扫描 2^20 个候选状态。这些调用彼此独立，窗口因此可以并行扫描；同时奇偶校验可以在 64 位目标上用一条宽表达式完成。

## mfoc 侧改动（third_party/mfoc/patches/0002-add-high-performance-and-x64-modes.patch）

- 新增 `src/nfcx-perf.c` / `src/nfcx-perf.h`：探测窗口扫描器。`nfcx_scan_probes()` 按 worker 分片调用未修改的 `lfsr_recovery32()` / `lfsr_rollback_word()` / `crypto1_get_lfsr()`，各自收集候选密钥后按 worker 顺序合并；候选集合与串行扫描完全一致，只有跨 worker 的排序不同（后续 `uniqsort` 不依赖顺序）。POSIX 使用 pthread，Windows 使用 `CreateThread`。
- `-H`（high performance）：启用上述多核扫描；worker 数默认取在线 CPU，可用 `NFCX_MFOC_THREADS` 覆盖，上限 64。启动时打印实际 worker 数。
- `-X`（x64）：把三个字节的奇偶校验合并到一条 64 位表达式，内部使用 `__builtin_parity()`（64 位目标上映射为硬件 popcount）。32 位构建会在启动阶段拒绝该模式并退出。
- 未加 `-H`/`-X` 时行为不变：串行扫描走同一个扫描器的串行入口，等价于原来的内联循环。
- 构建系统：`src/Makefile.am` 拆出 `libmfocattack.a`（crapto1/crypto1/nfcx-perf）供 mfoc 与测试共用；`configure.ac` 增加 `AC_PROG_RANLIB` 与 `AC_SEARCH_LIBS(pthread_create)`（Windows 不需要额外库）。

## 测试

`tests/nfcx-perf-test.c`（`make check`）：

1. `nfcx_valid_nonce_wide()` 与上游 `valid_nonce()` 公式在 2,000,000 组随机输入上逐例比对一致。
2. 32 个探测任务下，并行扫描与串行扫描恢复出的候选密钥集合（排序后）完全相同。
3. 16 核开发机实测：32 个探测串行 18.0–18.3 s，16 worker 并行 1.68 s，加速 10.7–10.9×。

## NFCX 侧

- `internal/attack`：新增 `MFoCMode`（`balanced` / `high_performance` / `x64`），`MFoCInvocation.Mode` 决定是否追加 `-H` / `-X`；`Available()` 额外要求帮助文本声明 `[-H]` / `[-X]`，旧 runtime 会明确报错而不是静默降级。
- `internal/workflow`：`MFoCRequest.Mode`、`KeyRecoveryRequest.Mode` 透传到 Nested 阶段；`mfoc-hardnested` 未打补丁，始终使用 `balanced`。
- `app`：`KeyRecoveryStartRequestDTO.Mode` / `MFoCStartRequestDTO.Mode` 接受 `balanced`、`high_performance`、`x64`（兼容 `performance`、`amd64` 等别名），未知值直接报错。
- 前端：恢复确认弹窗增加“mfoc 模式”下拉框，选择写入 `localStorage`，启动日志显示当前模式。
- `scripts/toolchain/mfoc.sh`：按 `MFOC_PATCH_SERIES` 顺序应用补丁序列、逐个安装到 runtime/patches，并在 `verify` 阶段检查帮助文本包含 `[-H] [-X]`。

## 验收

- `patch -p1` 在干净的 mfoc-0.10.7 源码上依次应用 0001、0002，`autoreconf` + `configure` + `make` + `make check` 全部通过。
- `mfoc -h` 输出新增 `-H` / `-X` 说明；`mfoc -H -X -O x.mfd` 打印模式横幅后按预期以“No NFC device found”退出（本机无读卡器）。
- `go test ./internal/attack ./internal/workflow ./app` 通过。
