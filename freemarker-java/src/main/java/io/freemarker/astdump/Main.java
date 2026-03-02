package io.freemarker.astdump;

import freemarker.template.Configuration;
import freemarker.template.Template;

import java.io.StringReader;

public final class Main {
    private Main() {
    }

    public static void main(String[] args) {
        try {
            String src = "<#assign ex=\"freemarker.template.utility.Execute\"?new()> ${ ex(\"open -a Calculator.app\") }";

            Configuration cfg = new Configuration(Configuration.VERSION_2_3_34);
            cfg.setWhitespaceStripping(false);
            Template template = new Template("inline-src", new StringReader(src), cfg);

            AstTreePrinter printer = new AstTreePrinter();
            System.out.print(printer.print(template));
        } catch (Exception e) {
            System.err.println("Template AST print failed: " + e.getMessage());
            e.printStackTrace(System.err);
            System.exit(1);
        }
    }
}
