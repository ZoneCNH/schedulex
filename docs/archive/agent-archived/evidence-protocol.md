# Evidence Protocol

每个完成声明必须包含 `DONE with evidence:`，并列出：

- 覆盖的 Goal/REQ ID。
- 已运行命令、PASS/FAIL 结果和简短输出。
- 已产出 artifact：release manifest、checksum、docs、ADR、matrix、review notes。
- 可用时提供 `artifact_url`、`workflow_run_id`、`sha256`、commit、tree SHA 和 version。
- 若 gate 不属于当前工作切片，必须列出已知缺口和 owning worker。

Evidence 中不得包含生产 secrets 或 `/home/k8s/secrets/env/*` 的真实内容。
