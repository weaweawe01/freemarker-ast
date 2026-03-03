package risk

import "strings"

const (
	maliciousClassScore      = 100
	newInstanceScore         = 40
	objectConstructorCall    = 60
	unsafeMethodMatchScore   = 80
	commandEvidenceScore     = 20
	apiBuiltinScore          = 60
	filePathEvidenceScore    = 40
	apiFileReadChainScore    = 120
	obfuscatedClassNameScore = 20
	fullExploitChainScore    = 100
)

var maliciousClasses = map[string]struct{}{
	"freemarker.template.utility.execute":           {},
	"freemarker.template.utility.objectconstructor": {},
	"freemarker.template.utility.jythonruntime":     {},
	"java.lang.processbuilder":                      {},
	"java.lang.runtime":                             {},
	"java.lang.system":                              {},
	"java.lang.class":                               {},
	"java.lang.classloader":                         {},
	"java.lang.reflect.method":                      {},
	"java.lang.reflect.constructor":                 {},
	"java.lang.reflect.field":                       {},
	"java.lang.reflect.accessibleobject":            {},
}

var criticalMethodScores = map[string]int{
	"start":              60,
	"exec":               90,
	"invoke":             80,
	"setaccessible":      80,
	"forname":            80,
	"newinstance":        75,
	"exit":               70,
	"halt":               80,
	"load":               70,
	"loadlibrary":        70,
	"setsecuritymanager": 80,
	"setproperty":        70,
	"stop":               60,
	"suspend":            60,
	"resume":             60,
	"getclassloader":     60,
}

var executionMethodSet = map[string]struct{}{
	"start":              {},
	"exec":               {},
	"invoke":             {},
	"setaccessible":      {},
	"forname":            {},
	"newinstance":        {},
	"exit":               {},
	"halt":               {},
	"load":               {},
	"loadlibrary":        {},
	"setsecuritymanager": {},
	"setproperty":        {},
	"stop":               {},
	"suspend":            {},
	"resume":             {},
}

var commandIndicators = []string{
	"whoami",
	"cmd.exe",
	"powershell",
	"bash -c",
	"sh -c",
	"/bin/bash",
	"/bin/sh",
	"calc",
	"curl ",
	"wget ",
	"nc ",
	"ncat ",
}

var fileIndicators = []string{
	"file://",
	"/etc/passwd",
	"/etc/shadow",
	"/proc/self/environ",
	"/root/.ssh",
	"win.ini",
	"c://windows",
	"c:\\windows",
}

var sensitiveIOMethodScores = map[string]int{
	"getresourceasstream": 90,
	"getresource":         70,
	"openconnection":      70,
	"getinputstream":      80,
	"read":                30,
	"touri":               20,
	"tourl":               20,
	"create":              30,
}

func isMaliciousClass(className string) bool {
	_, ok := maliciousClasses[className]
	return ok
}

func isExecutionMethod(name string) bool {
	_, ok := executionMethodSet[strings.ToLower(name)]
	return ok
}
