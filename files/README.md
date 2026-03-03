# AST 回归使用说明

本目录用于验证 `freemarker-go` 与 `freemarker-java` 的 AST 输出一致性。

## 目录

```text
files/
  ast-consistency-plan.md   # 规划文档
  generate_large_cases.sh   # 批量生成大规模样例
  run_ast_regression.sh     # 一键回归脚本
  cases/
    basic/                  # 基础语句
    complex/                # 复杂语句
    nested/                 # 嵌套语句
  out/
    java/                   # Java 输出 AST
    go/                     # Go 输出 AST
  diff/                     # 差异文件（仅失败用例）
```

## 前置条件

1. Go 1.22+。
2. JDK 8+。
3. Maven（`mvn` 可执行）。

## 执行

在 `freemarker-go` 目录执行：

```bash
cd /Users/cy01/Desktop/代码仓库/freemarker-go/src/freemarker-go
bash files/run_ast_regression.sh
```

如果 `mvn` 不在 `PATH`，可传入 Maven 可执行文件路径：

```bash
MVN_BIN="/Applications/IntelliJ IDEA.app/Contents/plugins/maven/lib/maven3/bin/mvn" \
  bash files/run_ast_regression.sh
```

可按阶段执行：

```bash
bash files/run_ast_regression.sh basic
bash files/run_ast_regression.sh complex
bash files/run_ast_regression.sh nested
```

脚本行为：

1. Java 侧批量读取 `cases/*.ftl` 输出到 `files/out/java/**/*.ast`。
2. Go 侧批量读取同一文件输出到 `files/out/go/**/*.ast`。
3. 统一归一化后比较 AST。
4. 不一致时生成 `files/diff/**/*.diff`。

## 生成大样本

默认会生成每类 400 条（`basic/complex/nested`），总计 1200+ 条：

```bash
cd /Users/cy01/Desktop/代码仓库/freemarker-go/src/freemarker-go
bash files/generate_large_cases.sh
```

也可自定义数量（参数顺序：basic complex nested）：

```bash
bash files/generate_large_cases.sh 500 500 500
```

## 新增用例

1. 在对应阶段目录新增 `.ftl` 文件，命名建议：
   `Bxxx_*.ftl`、`Cxxx_*.ftl`、`Nxxx_*.ftl`。
2. 重新执行脚本。
3. 若失败，查看 `files/diff/*.diff` 定位第一处差异。
