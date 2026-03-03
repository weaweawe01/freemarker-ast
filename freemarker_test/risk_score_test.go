package freemarker_test

import (
	"fmt"
	"testing"

	freemarker "github.com/weaweawe01/freemarker-ast"
)

func TestAnalyzeRisk_ObjectConstructorProcessBuilder(t *testing.T) {
	src := `<#assign value="freemarker.template.utility.ObjectConstructor"?new()>${value("java.lang.ProcessBuilder","whoami").start()}`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	fmt.Println(report.TotalScore)
	if report.TotalScore < 400 {
		t.Fatalf("expected high risk score >= 400, got %d", report.TotalScore)
	}
	if report.Severity != freemarker.RiskSeverity("critical") {
		t.Fatalf("expected critical severity, got %q", report.Severity)
	}
	if !containsRule(report.Findings, "MALICIOUS_CLASS") {
		t.Fatalf("expected MALICIOUS_CLASS finding, got %#v", report.Findings)
	}
	if !containsRule(report.Findings, "FULL_EXPLOIT_CHAIN") {
		t.Fatalf("expected FULL_EXPLOIT_CHAIN finding, got %#v", report.Findings)
	}
}

func TestAnalyzeRisk_ExecuteCall(t *testing.T) {
	src := `<#assign value="freemarker.template.utility.Execute"?new()>${value("cmd.exe /c calc")}`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	if report.TotalScore < 300 {
		t.Fatalf("expected score >= 300, got %d", report.TotalScore)
	}
	if !containsRule(report.Findings, "EXECUTE_CALL") {
		t.Fatalf("expected EXECUTE_CALL finding, got %#v", report.Findings)
	}
}

func TestAnalyzeRisk_UnsafeMethodsBlacklistHit(t *testing.T) {
	src := `<#assign value="freemarker.template.utility.ObjectConstructor"?new()>${value("java.lang.Runtime").exec("whoami")}`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	if !containsRule(report.Findings, "UNSAFE_METHOD_BLACKLIST") {
		t.Fatalf("expected UNSAFE_METHOD_BLACKLIST finding, got %#v", report.Findings)
	}
	if report.TotalScore < 500 {
		t.Fatalf("expected strong score >= 500, got %d", report.TotalScore)
	}
}

func TestAnalyzeRisk_BenignTemplate(t *testing.T) {
	src := `<#assign x = 1>${x}`
	report, err := freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	if report.TotalScore != 0 {
		t.Fatalf("expected score 0 for benign template, got %d", report.TotalScore)
	}
	if report.Severity != freemarker.RiskSeverity("low") {
		t.Fatalf("expected low severity, got %q", report.Severity)
	}
}

func TestAnalyzeRisk_MaliciousClassAtLeast100(t *testing.T) {
	src := `${"freemarker.template.utility.Execute"}`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	if report.TotalScore < 100 {
		t.Fatalf("expected score >= 100 when malicious class is referenced, got %d", report.TotalScore)
	}
	if !containsRule(report.Findings, "MALICIOUS_CLASS") {
		t.Fatalf("expected MALICIOUS_CLASS finding, got %#v", report.Findings)
	}
}

func TestAnalyzeRisk_ApiResourceRead_WinIni(t *testing.T) {
	src := `<#assign is=object?api.class.getResourceAsStream("c://windows/win.ini")>`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	if report.TotalScore <= 0 {
		t.Fatalf("expected score > 0, got %d", report.TotalScore)
	}
	if !containsRule(report.Findings, "API_BUILTIN") {
		t.Fatalf("expected API_BUILTIN finding, got %#v", report.Findings)
	}
	if !containsRule(report.Findings, "SENSITIVE_IO_METHOD") {
		t.Fatalf("expected SENSITIVE_IO_METHOD finding, got %#v", report.Findings)
	}
	if !containsRule(report.Findings, "FILE_PATH_EVIDENCE") {
		t.Fatalf("expected FILE_PATH_EVIDENCE finding, got %#v", report.Findings)
	}
}

func TestAnalyzeRisk_ApiResourceRead_PasswdChain(t *testing.T) {
	src := `<#assign uri=object?api.class.getResource("/").toURI()>
<#assign input=uri?api.create("file:///etc/passwd").toURL().openConnection()>
<#assign is=input?api.getInputStream()>
FILE:[<#list 0..999999999 as _>
    <#assign byte=is.read()>
    <#if byte == -1>
        <#break>
    </#if>
${byte}, </#list>]`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	if report.TotalScore < 200 {
		t.Fatalf("expected strong score >= 200, got %d", report.TotalScore)
	}
	if !containsRule(report.Findings, "API_FILE_READ_CHAIN") {
		t.Fatalf("expected API_FILE_READ_CHAIN finding, got %#v", report.Findings)
	}
}

func TestAnalyzeRisk_ApiResourceRead_PasswdChain22(t *testing.T) {
	src := `<#assign is=object?api.class.getResourceAsStream("/Test.class")>
FILE:[<#list 0..999999999 as _>
    <#assign byte=is.read()>
    <#if byte == -1>
        <#break>
    </#if>
${byte}, </#list>]
<#assign uri=object?api.class.getResource("/").toURI()>
<#assign input=uri?api.create("file:///etc/passwd").toURL().openConnection()>
<#assign is=input?api.getInputStream()>
FILE:[<#list 0..999999999 as _>
    <#assign byte=is.read()>
    <#if byte == -1>
        <#break>
    </#if>
${byte}, </#list>]`
	report, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		t.Fatalf("AnalyzeRisk returned error: %v", err)
	}
	fmt.Println(report.TotalScore)
}
func TestAnalyzeRisk_MaliciousClassAtLeast1002(t *testing.T) {
	src := `<#assign value="freemarker.template.utility.Execute"?new()>${value("calc")}`
	report, _ := freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `<#assign ex="freemarker.template.utility.Execute"?new()> ${ ex("id") }`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `${"freemarker.template.utility.Execute"?new()("calc")}`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `<#assign ex="freemarker.template.utility.Execute"?new()>${ex("id")}`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `<html>
<head>
    <title>User Info</title>
</head>
<body>
<h1>Hello, ${user.username}!</h1>
<#assign value="freemarker.template.utility.Execute"?new()>${value("cmd.exe /c type C:\\Windows\\win.ini")}
</body>
</html>`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `<#assign value="freemarker.template.utility.JythonRuntime"?new()><@value>import os;os.system("cmd.exe /c calc")</@value>`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)
	src = `<#assign value="freemarker.template.utility.ObjectConstructor"?new()>${value("java.lang.ProcessBuilder","cmd.exe","/c","calc").start()}`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `
        <#assign is=object?api.class.getResourceAsStream("c://windows/win.ini")>
        `
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `
        <#assign uri=object?api.class.getResource("/").toURI()>
<#assign input=uri?api.create("file:///etc/passwd").toURL().openConnection()>
<#assign is=input?api.getInputStream()>
FILE:[<#list 0..999999999 as _>
    <#assign byte=is.read()>
    <#if byte == -1>
        <#break>
    </#if>
${byte}, </#list>]
        `
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

	src = `<#assign classLoader=object?api.class.protectionDomain.classLoader> 
<#assign clazz=classLoader.loadClass("ClassExposingGSON")> 
<#assign field=clazz?api.getField("GSON")> 
<#assign gson=field?api.get(null)> 
<#assign ex=gson?api.fromJson("{}", classLoader.loadClass("freemarker.template.utility.Execute"))> 
${ex("calc")}`
	report, _ = freemarker.AnalyzeRisk(src)
	fmt.Println(report.TotalScore)

}

func containsRule(findings []freemarker.RiskFinding, rule string) bool {
	for _, f := range findings {
		if f.Rule == rule {
			return true
		}
	}
	return false
}
