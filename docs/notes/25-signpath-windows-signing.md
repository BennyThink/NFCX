# SignPath Windows 发布签名

## 实际完成

- Windows release job 改为通过 SignPath GitHub Action 提交签名请求，不再向 CI 导入 Windows PFX 私钥；
- 使用项目 `NFCX`、测试签名策略 `test-signing` 和 Artifact Configuration `nfcx-windows-test-sign`；API token 继续仅从 `SIGNPATH_API_TOKEN` repository secret 读取；
- 先生成临时 unsigned ZIP，SignPath 仅对其明确配置的 EXE/DLL 签名，再将取回的已签名文件放回 Windows staging 目录；
- 签名后才生成 runtime manifest，因此 manifest 的 SHA-256 与最终发布的 Authenticode-signed 文件一致；
- 最终 GitHub Release 仅上传签名后的 `NFCX-<version>-windows-amd64.zip`，临时 `-unsigned` ZIP 不会作为发布 asset 保留；
- workflow 在打包前检查每个分发的 EXE/DLL 是否含嵌入式签名。

## 实现中遇到的问题

- 原 workflow 仅支持从 `WINDOWS_CERTIFICATE` 和 `WINDOWS_CERTIFICATE_PASSWORD` 导入 PFX。SignPath 托管证书不应导出私钥到 GitHub runner，因此该路径已由 SignPath signing request 取代。
- runtime manifest 必须在 Authenticode 写入 PE 文件之后生成；若提前生成，签名会改变文件哈希，应用的 `--self-check` 会错误报告完整性失败。

## 阻塞或未完成

- 无。下一次带 tag 的 GitHub Actions release 将使用 SignPath 提供的 self-signed 测试证书验证完整端到端流程；获得生产证书后，只需在 SignPath 中切换 signing policy slug。
