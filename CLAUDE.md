# CLAUDE.md

> 完整的贡献指南、测试规范、提交规范和 Agent 协作约定见 [AGENTS.md](AGENTS.md)。

## 语言规则（全局强制）

1. **回答语言**：所有对话回复默认使用中文，除非用户明确要求使用其他语言。
2. **文档语言**：所有仓库文档（README、docs/、.agent/、contracts/*.md、变更日志、发布说明、PR 描述、Issue、贡献指南）默认使用中文叙述。
3. **代码注释**：Go 源码中的注释（包括函数文档注释、行内注释、TODO/FIXME）默认使用中文。导出符号的 godoc 注释若面向外部消费者可保留英文，内部代码一律中文。
4. **保留原文的例外**：代码标识符、命令、路径、包名、Go module 路径、外部专有名词（Agent、Harness、manifest、schema、CI、PR、Issue）、协议固定短语和 git 提交标题保留项目惯用原文。
5. **提交信息**：提交正文（body）和 trailer 使用中文；提交标题（subject line）保留英文以兼容工具链。
