# FreeMarker-Go 与 Java AST 一致性验证规划

## 1. 目标

验证 `freemarker-go` 对同一份 `.ftl` 模板的 AST 输出，是否与 `freemarker-go/freemarker-java`（Java 参考实现）一致，并形成可重复执行的回归测试流程。

核心原则：

1. 同一输入文件（同一份 `.ftl`）。
2. 分别由 Java 和 Go 产出 AST。
3. 归一化后进行逐行对比。
4. 先基础语句，再复杂语句，最后多层嵌套语句。

## 2. 范围与约束

范围：

1. 仅覆盖 AST 结构一致性（节点类型、参数、层级、顺序、值）。
2. 不覆盖模板渲染结果一致性（运行时执行输出）。
3. 测试样例统一放在 `freemarker-go/files` 目录。

约束：

1. Java 侧当前入口是固定模板字符串，需先支持“按文件输入输出 AST”。
2. Go 侧当前也建议增加“按文件输入输出 Java 风格 AST”的命令入口，便于批量回归。

## 3. 目录规划

在 `freemarker-go/files` 下建立如下结构：

```text
freemarker-go/files/
  README.md
  cases/
    basic/
    complex/
    nested/
  out/
    java/
    go/
  diff/
```

说明：

1. `cases/*`：测试输入 `.ftl` 文件。
2. `out/java/*`：Java 生成的 AST（基准）。
3. `out/go/*`：Go 生成的 AST（待比对）。
4. `diff/*`：差异报告（仅当不一致时生成）。

建议命名规范：

1. `B001_...ftl`：基础语句。
2. `C001_...ftl`：复杂语句。
3. `N001_...ftl`：嵌套语句。

## 4. 单用例验证流程

对每个 `.ftl` 用例执行以下步骤：

1. Java 解析并导出 AST 到 `out/java/<case>.ast`。
2. Go 解析并导出 AST 到 `out/go/<case>.ast`。
3. 归一化两个 AST 文本（换行符、BOM、头部注释、尾随空行）。
4. 按行严格比较。
5. 不一致时输出首个差异行与完整 diff 到 `diff/<case>.diff`。
6. 记录结果到总报告（PASS/FAIL）。

归一化规则建议与当前 `freemarker-go/freemarker_test/helpers_test.go` 对齐：

1. `\r\n` 和 `\r` 统一成 `\n`。
2. 去掉 BOM。
3. 去掉 AST 头部版权注释块（如果存在）。
4. 末尾空行只保留统一形式（或全部去掉）。

## 5. 分阶段测试清单

### Phase 1：基础语句（Basic）

目标：验证解析主干路径，尽快建立第一批“稳定通过”的回归样例。

建议样例：

1. `B001_text_only`：纯文本。
2. `B002_interpolation`：`${x}`。
3. `B003_assign`：`<#assign x = 1>`。
4. `B004_if_else`：`<#if x>...<#else>...</#if>`。
5. `B005_list`：`<#list xs as x>...</#list>`。
6. `B006_comment`：`<#-- comment -->`。
7. `B007_macro_simple`：`<#macro m>...</#macro>`。

通过标准：

1. Basic 阶段所有用例 Java/Go AST 完全一致。
2. 如有差异，先修复解析或打印器，再进入下一阶段。

### Phase 2：复杂语句（Complex）

目标：覆盖表达式和参数体系中更高复杂度分支。

建议样例：

1. `C001_builtin_chain`：`x?trim?upper_case`。
2. `C002_range_expr`：`1..10`、`1..<10`、`1..!10`。
3. `C003_hash_and_list_literal`：`{...}`、`[...]`。
4. `C004_default_exists`：`x!\"d\"`、`x??`。
5. `C005_call_and_method`：方法调用和参数列表。
6. `C006_lambda_like`：`x -> ...`（若语法版本支持）。
7. `C007_whitespace_stripping`：空白裁剪相关语法。

通过标准：

1. Complex 阶段全部一致。
2. 每个差异至少定位到“词法/语法树构建/AST打印”中的一个具体责任点。

### Phase 3：嵌套语句（Nested）

目标：验证深层结构和组合路径，覆盖最容易出现层级偏差的问题。

建议样例：

1. `N001_if_in_list`：`list` 内嵌 `if`。
2. `N002_list_in_if`：`if` 分支内嵌 `list`。
3. `N003_nested_macro_calls`：宏定义 + 宏调用嵌套。
4. `N004_mixed_directives`：`assign + if + list + interpolation` 混合。
5. `N005_deep_nesting`：3 层以上复合嵌套。
6. `N006_nested_ignored_children`：覆盖已知忽略子节点场景。

通过标准：

1. Nested 阶段全部一致。
2. 对历史问题场景保留回归用例，避免后续退化。

## 6. 执行节奏建议

1. 第 1 天：补齐 Java/Go 文件输入输出入口，跑通 1 个 Basic 用例。
2. 第 2 天：完成 Basic 全量，修复首轮差异。
3. 第 3 天：完成 Complex 全量，沉淀差异分类。
4. 第 4 天：完成 Nested 全量，形成阶段报告与回归脚本。

## 7. 产出物

1. `freemarker-go/files/cases/*`：分阶段 `.ftl` 测试样例。
2. `freemarker-go/files/out/java/*`：Java AST 基准输出。
3. `freemarker-go/files/out/go/*`：Go AST 实际输出。
4. `freemarker-go/files/diff/*`：不一致用例差异文件。
5. `freemarker-go/files/README.md`：如何执行与如何新增用例。

## 8. 验收标准

1. 三个阶段（Basic/Complex/Nested）所有用例均通过 AST 一致性对比。
2. 差异报告可定位到具体 case、具体行号、具体字段。
3. 回归流程可一键执行，新增 case 不需要改代码逻辑。
4. 在 CI 或本地重复执行结果一致、稳定。

## 9. 后续扩展（可选）

1. 将 `ast/core` 现有官方样例批量纳入 `files/cases` 回归集。
2. 在 PR 流程里自动执行 AST 对比，失败即阻断合并。
3. 为高风险语法（如 lambda、复杂内置链、深层嵌套）设置单独的回归标签。
