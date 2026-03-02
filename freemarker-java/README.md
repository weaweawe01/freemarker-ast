# freemarker-java

这是一个独立的 Java（Maven）项目，用于把 FreeMarker 模板解析成 AST，并按 `ast-1.ast` 同风格输出。

## 当前目标

- 输出格式尽量对齐 `test/resources/freemarker/core/*.ast`
- 支持基线对比：`ast-1.ftl` 生成结果与 `ast-1.ast` 对比

## 目录结构

```text
freemarker-java/
  pom.xml
  src/main/java/io/freemarker/astdump/
    Main.java                   # 固定模板输出 AST（.ast 风格）
    AstTreePrinter.java         # .ast 风格打印器
    AstFixtureCompareMain.java  # 与 ast-1.ast 做一致性对比
    AstDumper.java              # 旧版 JSON 导出器（保留）
```

## 输出风格示例

```text
#mixed_content  // f.c.MixedContent
    #assign  // f.c.Assignment
        - assignment target: "ex"  // String
        - assignment operator: "="  // String
        - assignment source: ...(...)  // f.c.MethodCall
```

## 运行方式

### 1. 运行固定模板 AST 输出（`Main`）

```bash
cd freemarker-java
mvn -q exec:java
```

当前 `Main` 内固定模板为：

```ftl
<#assign ex="freemarker.template.utility.Execute"?new()> ${ ex("open -a Calculator.app") }
```

### 2. 运行 `ast-1` 基线对比

`AstFixtureCompareMain` 会：
- 读取 `../test/resources/freemarker/core/ast-1.ftl`
- 按测试方式去掉头部版权注释并规范行尾
- 输出 AST（.ast 风格）
- 与 `../test/resources/freemarker/core/ast-1.ast` 的主体内容比较

输出为：
- `MATCH: ...` 表示一致
- `DIFF: ...` 表示不一致，并打印首批差异行

若你本地没装 Maven，也可以直接 `javac/java` 跑。

## 说明

- 这个项目依赖 FreeMarker 内部 AST API（`freemarker.core.TemplateObject` 等），参数读取使用反射。
- 由于是内部 API，FreeMarker 升级后可能需要调整。
