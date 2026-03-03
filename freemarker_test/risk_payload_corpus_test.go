package freemarker_test

import (
	"fmt"
	"testing"

	freemarker "github.com/weaweawe01/freemarker-ast"
)

func TestAnalyzeRisk_DangerousPayloadCorpus100(t *testing.T) {
	payloads := buildDangerousPayloadCorpus()
	if len(payloads) != 100 {
		t.Fatalf("expected 100 dangerous payloads, got %d", len(payloads))
	}

	for i, src := range payloads {
		report, err := freemarker.AnalyzeRisk(src)
		if err != nil {
			t.Fatalf("dangerous payload[%d] parse/analyze failed: %v\nsrc=%s", i, err, src)
		}
		if report.TotalScore <= 0 {
			t.Fatalf("dangerous payload[%d] expected score > 0, got %d\nsrc=%s", i, report.TotalScore, src)
		}
	}
}

func TestAnalyzeRisk_BenignPayloadCorpus100(t *testing.T) {
	payloads := buildBenignPayloadCorpus()
	if len(payloads) != 100 {
		t.Fatalf("expected 100 benign payloads, got %d", len(payloads))
	}

	for i, src := range payloads {
		report, err := freemarker.AnalyzeRisk(src)
		if err != nil {
			t.Fatalf("benign payload[%d] parse/analyze failed: %v\nsrc=%s", i, err, src)
		}
		if report.TotalScore != 0 {
			t.Fatalf("benign payload[%d] expected score 0, got %d\nsrc=%s", i, report.TotalScore, src)
		}
	}
}

func buildDangerousPayloadCorpus() []string {
	out := make([]string, 0, 100)

	for i := 0; i < 25; i++ {
		out = append(out,
			fmt.Sprintf(`<#assign ex%d="freemarker.template.utility.Execute"?new()>${ex%d("whoami")}`, i, i),
		)
	}

	for i := 0; i < 25; i++ {
		out = append(out,
			fmt.Sprintf(
				`<#assign oc%d="freemarker.template.utility.ObjectConstructor"?new()>${oc%d("java.lang.ProcessBuilder","cmd.exe","/c","echo %d").start()}`,
				i, i, i,
			),
		)
	}

	for i := 0; i < 25; i++ {
		out = append(out,
			fmt.Sprintf(
				`<#assign rt%d="freemarker.template.utility.ObjectConstructor"?new()>${rt%d("java.lang.Runtime").exec("whoami")}`,
				i, i,
			),
		)
	}

	for i := 0; i < 25; i++ {
		out = append(out,
			fmt.Sprintf(
				`<#assign is%d=obj?api.class.getResourceAsStream("c://windows/win.ini")>`,
				i,
			),
		)
	}

	return out
}

func buildBenignPayloadCorpus() []string {
	out := make([]string, 0, 100)

	for i := 0; i < 40; i++ {
		out = append(out, fmt.Sprintf(`<#assign x%d = %d>${x%d}`, i, i, i))
	}

	for i := 0; i < 30; i++ {
		out = append(out, fmt.Sprintf(`<#if %d gt 0>safe_%d</#if>`, i+1, i))
	}

	for i := 0; i < 20; i++ {
		out = append(out, fmt.Sprintf(`<#list [1,2,3] as item%d>${item%d}</#list>`, i, i))
	}

	for i := 0; i < 10; i++ {
		out = append(out, fmt.Sprintf(`${"hello_%d"?upper_case}`, i))
	}

	return out
}
