# freemarker-ast

[Apache FreeMarker](https://freemarker.apache.org/) 模板解析器的 Go 语言实现，生成的 AST 与 Java 参考实现 **100% 兼容**。

## 项目简介

`freemarker-ast` 解析 FreeMarker 模板源文件（`.ftl`），并以与 Java `freemarker-core` 库完全相同的文本格式输出抽象语法树（AST）。每个节点类型、类名、字段标签及结构细节均与 Java 输出精确一致，可作为 Go 工具链中的解析后端直接使用。


## 安装

```bash
go get github.com/weaweawe01/freemarker-ast
```


## 快速开始

### 解析模板并输出 AST

```go
package main

import (
    "fmt"
    "log"

    freemarker "github.com/weaweawe01/freemarker-ast"
)

func main() {
    src := `<#assign x = "hello">${x?upper_case}`
    out, err := freemarker.ParseToJavaLikeAST(src)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(out)
}
```

输出结果（与 Java FreeMarker 完全相同）：

```
#mixed_content  // f.c.MixedContent
    #assign  // f.c.Assignment
        - assignment target: "x"  // String
        - assignment operator: "="  // String
        - assignment source: "hello"  // f.c.StringLiteral
        - variable scope: "1"  // Integer
        - namespace: null  // Null
    ${...}  // f.c.DollarVariable
        - content: ?upper_case  // f.c.BuiltInsForStringsBasic$upper_caseBI
            - left-hand operand: x  // f.c.Identifier
            - right-hand operand: "upper_case"  // String
```


### AST 恶意打分 AST

```go
package main

import (
	"fmt"
	"os"

	"github.com/weaweawe01/freemarker-ast"
)

func main() {
	src := `<#assign ex="freemarker.template.utility.Execute"?new()> ${ ex("id") }`
	out, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out.TotalScore)

}

```

```shell
root@hcss-ecs-5ed3:~/goggg# go run main.go 
330

```



## 目录结构

```
github.com/weaweawe01/freemarker-ast/
├── internal/
│   ├── lexer/        # FreeMarker 词法分析器
│   ├── parser/       # 递归下降解析器 → ast.Root
│   ├── ast/          # AST 节点模型
│   ├── astdump/      # AST → Java 兼容文本格式
│   ├── tokenid/      # Token 类型常量（自动生成）
│   ├── tokenspec/    # Token 规格辅助工具
│   ├── corpus/       # 测试语料库发现工具
│   ├── oracle/       # Oracle 基准引导与差异对比工具
│   ├── compat/       # 兼容性类型辅助
│   └── diff/         # JSON 差异对比工具
├── cmd/
│   ├── fm-core-parse-check/   # CLI：解析检查模板文件
│   ├── fm-oracle-bootstrap/   # CLI：引导 Oracle fixture
│   ├── fm-oracle-diff/        # CLI：对比 Oracle 结果
│   ├── fm-token-gen/          # CLI：生成 Token 常量
│   └── fm-token-spec/         # CLI：查看 Token 规格
├── freemarker_test/  # 与 Java AST fixture 的奇偶校验测试
└── ast/core/         # AST + FTL fixture 文件（基准来源）
```

## 参考实现

本项目基于 **[Apache FreeMarker 2.3.34](https://freemarker.apache.org/)** 进行重构，所有 AST 节点结构、类名、字段命名均以该版本的 Java 源码为准。

### freemarker-java（AST 基准生成工具）

仓库中的 `freemarker-java/` 子目录是一个独立的 Java 工具项目（`freemarker-java-ast-dump`），其作用是：

- 引入 `org.freemarker:freemarker:2.3.34` 官方依赖
- 解析 `.ftl` 模板并将完整 AST 以 JSON 格式导出
- 导出结果作为 `ast/core/*.ast` fixture 文件的生成来源，是本 Go 实现奇偶校验测试的**权威基准**

| 属性 | 值 |
|---|---|
| 基准 Java 库版本 | `org.freemarker:freemarker:2.3.34` |
| 工具 artifactId | `freemarker-java-ast-dump` |
| Jackson 版本 | `2.18.2` |
| 编译目标 | Java 1.8+ |

只要 `ast/core/*.ast` fixture 文件与 Java 工具的输出完全一致，Go 实现即视为与 Java 参考实现 100% 兼容。

## AST 兼容性

AST 输出通过 [`ast/core/`](ast/core/) 目录中的 fixture 文件进行验证，这些文件由 Java `freemarker-core` 库生成。每组 fixture 包含：

- `<caseName>.ftl` — FreeMarker 模板源文件
- `<caseName>.ast` — Java 文本格式的期望 AST 输出

当前 fixture 覆盖情况：

| Fixture | 说明 |
|---|---|
| `ast-1` | 通用指令与表达式 |
| `ast-assignments` | 变量赋值形式 |
| `ast-builtins` | 内置运算符表达式（`?`） |
| `ast-lambda` | Lambda 表达式（`->`） |
| `ast-locations` | 源码位置信息 |
| `ast-mixedcontentsimplifications` | 混合内容简化 |
| `ast-multipleignoredchildren` | 多个忽略子节点 |
| `ast-nestedignoredchildren` | 嵌套忽略子节点 |
| `ast-range` | 范围表达式（`..`、`..<`、`..*`、`..!`） |
| `ast-strlitinterpolation` | 字符串字面量插值 |
| `ast-whitespacestripping` | 空白符剥除规则 |

## AST 一致性回归（files）

仓库新增 `files/` 目录用于按阶段（`basic/complex/nested`）执行 Java/Go AST 对比回归：

```bash
cd freemarker-go
bash files/run_ast_regression.sh
```

支持按阶段执行：

```bash
bash files/run_ast_regression.sh basic
bash files/run_ast_regression.sh complex
bash files/run_ast_regression.sh nested
```

脚本会将输出分别写入：

1. `files/out/java/*.ast`（Java 基准输出）
2. `files/out/go/*.ast`（Go 实际输出）
3. `files/diff/*.diff`（不一致时生成）
