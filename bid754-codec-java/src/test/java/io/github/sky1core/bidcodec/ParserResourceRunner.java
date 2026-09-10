package io.github.sky1core.bidcodec;

import com.sun.management.ThreadMXBean;
import java.lang.management.ManagementFactory;
import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;

public final class ParserResourceRunner {
    private record Width(int bits, int precision, int qmin, int qmax) {}
    private record ParseCase(String name, String input, Components expected, String errorPart) {}
    private static final Width[] WIDTHS = {
        new Width(32, 7, -101, 90), new Width(64, 16, -398, 369), new Width(128, 34, -6176, 6111)
    };
    private static final int[] LENGTHS = {34, 35, 6176, 6177, 61760};
    private static final int ALLOCATION_BUDGET = 4096;
    private static final int WARMUP = 128;
    private static final int REPETITIONS = 32;
    private static volatile Object sink;
    private static int cases;

    private static void check(boolean condition, String label) {
        if (!condition) throw new AssertionError(label);
    }

    private static BigInteger magnitude(BigInteger value) {
        return value == null ? BigInteger.ZERO : value;
    }

    private static void fields(Components got, Components want, String label) {
        check(got.sign() == want.sign() && got.kind() == want.kind()
                && got.exponent() == want.exponent()
                && magnitude(got.coefficient()).equals(magnitude(want.coefficient()))
                && magnitude(got.payload()).equals(magnitude(want.payload())), label + ": fields");
    }

    private static void boundedError(IllegalArgumentException error, String label) {
        check(error.getMessage() != null && error.getMessage().getBytes(StandardCharsets.UTF_8).length <= 256,
                label + ": error exceeds 256 bytes");
    }

    private static void parseCase(ParseCase test) {
        Components got;
        try {
            got = BidCodec.fromString(test.input());
        } catch (IllegalArgumentException error) {
            check(test.expected() == null, test.name() + ": unexpected rejection");
            boundedError(error, test.name());
            check(error.getMessage().contains(test.errorPart()), test.name() + ": wrong rejection boundary");
            cases++;
            return;
        }
        check(test.expected() != null, test.name() + ": accepted invalid input");
        fields(got, test.expected(), test.name());
        fields(BidCodec.fromString(BidCodec.toString(got)), test.expected(), test.name() + ": cohort closure");
        cases++;
    }

    private static Components normal(String digits, int exponent) {
        return new Components(false, new BigInteger(digits), exponent);
    }

    private static Components roundTrip(Width width, Components value) {
        return switch (width.bits()) {
            case 32 -> BidCodec.decode32(BidCodec.encode32(value));
            case 64 -> BidCodec.decode64(BidCodec.encode64(value));
            case 128 -> {
                long[] words = BidCodec.encode128(value);
                yield BidCodec.decode128(words[0], words[1]);
            }
            default -> throw new AssertionError("unknown width");
        };
    }

    private static void widthCase(Width width, Components value, boolean valid, String name) {
        try {
            Components decoded = roundTrip(width, value);
            check(valid, name + ": encoded out-of-width input");
            fields(decoded, value, name);
        } catch (IllegalArgumentException error) {
            check(!valid, name + ": rejected representable value");
            boundedError(error, name);
        }
        cases++;
    }

    private static void widthBoundaries() {
        for (Width width : WIDTHS) {
            for (int digits : new int[]{width.precision() - 1, width.precision(), width.precision() + 1}) {
                for (int exponent : new int[]{width.qmin() - 1, width.qmin(), width.qmax(), width.qmax() + 1}) {
                    String coefficient = "9".repeat(digits);
                    String name = "bid" + width.bits() + " digits=" + digits + " q=" + exponent;
                    Components value = normal(coefficient, exponent);
                    parseCase(new ParseCase(name, coefficient + "e" + exponent, digits <= 34 ? value : null, ""));
                    widthCase(width, value, digits <= width.precision() && exponent >= width.qmin()
                            && exponent <= width.qmax(), name);
                }
            }
            for (int exponent : new int[]{width.qmin() - 1, width.qmin(), width.qmax(), width.qmax() + 1}) {
                Components zero = new Components(true, exponent, DecimalKind.ZERO);
                parseCase(new ParseCase("signed zero", "-0e" + exponent, zero, ""));
                widthCase(width, zero, exponent >= width.qmin() && exponent <= width.qmax(), "zero width");
            }
            for (int digits : new int[]{width.precision() - 2, width.precision() - 1, width.precision()}) {
                String payload = "9".repeat(digits);
                Components value = new Components(true, DecimalKind.SNAN, new BigInteger(payload));
                parseCase(new ParseCase("payload schema", "-sNaN" + payload, digits <= 33 ? value : null, ""));
                widthCase(width, value, digits < width.precision(), "payload width");
            }
        }
    }

    private static List<ParseCase> parserCases() {
        List<ParseCase> tests = new ArrayList<>();
        tests.add(new ParseCase("schema34", "9".repeat(34), normal("9".repeat(34), 0), ""));
        tests.add(new ParseCase("schema35", "9".repeat(35), null, ""));
        tests.add(new ParseCase("payload33", "NaN" + "9".repeat(33),
                new Components(false, DecimalKind.QNAN, new BigInteger("9".repeat(33))), ""));
        tests.add(new ParseCase("payload34", "NaN" + "9".repeat(34), null, ""));
        for (long q : new long[]{-2147483649L, -2147483648L, -2147483647L, 2147483646L, 2147483647L, 2147483648L}) {
            tests.add(new ParseCase("int32 " + q, "1e" + q,
                    q >= Integer.MIN_VALUE && q <= Integer.MAX_VALUE ? normal("1", (int) q) : null, ""));
        }
        for (String sign : new String[]{"", "-"}) {
            tests.add(new ParseCase("literal inside", "1e" + sign + "9007199254740991", null, "32-bit"));
            tests.add(new ParseCase("literal at bound", "1e" + sign + "9007199254740992", null, "2^53"));
            tests.add(new ParseCase("literal outside", "1e" + sign + "9007199254740993", null, "2^53"));
        }
        tests.add(new ParseCase("fraction upper inside", "1.0e2147483648", normal("10", Integer.MAX_VALUE), ""));
        tests.add(new ParseCase("fraction upper outside", "1.0e2147483649", null, ""));
        tests.add(new ParseCase("fraction lower inside", "1.0e-2147483647", normal("10", Integer.MIN_VALUE), ""));
        tests.add(new ParseCase("fraction lower outside", "1.0e-2147483648", null, ""));
        tests.add(new ParseCase("cohort", "-0001.100", new Components(true, new BigInteger("1100"), -3), ""));
        tests.add(new ParseCase("trim6", "\t\n\u000b\f\r -iNfInItY \t\n\u000b\f\r", new Components(true, DecimalKind.INFINITY), ""));
        tests.add(new ParseCase("negative exponent zero", "-0e-0", new Components(true, 0, DecimalKind.ZERO), ""));
        for (String s : new String[]{"", "+", "-", ".", "e1", "1e", "1e+", "1e-", "1e+-1", "1.2.3", "1e1.0",
                "1 0", "NaN+1", "SNaN-1", "Infinityx", "\u00001", "1\u001f", "1_0", "0x10", "\u00a01", "１", "1\u007f", "1\n0"}) {
            tests.add(new ParseCase("grammar", s, null, ""));
        }
        for (int n : LENGTHS) {
            String zeros = "0".repeat(n);
            tests.add(new ParseCase("leading coefficient " + n, " -" + zeros + "1200 ", new Components(true, new BigInteger("1200"), 0), ""));
            tests.add(new ParseCase("leading payload " + n, " -sNaN" + zeros + "12 ", new Components(true, DecimalKind.SNAN, BigInteger.valueOf(12)), ""));
            tests.add(new ParseCase("leading exponent " + n, "1e-" + zeros + "12", normal("1", -12), ""));
            tests.add(new ParseCase("leading fraction " + n, "0." + zeros + "1200e" + (n + 4), normal("1200", 0), ""));
            tests.add(new ParseCase("allzero " + n, "-" + zeros, new Components(true, 0, DecimalKind.ZERO), ""));
            tests.add(new ParseCase("fractionzero " + n, "-0." + zeros, new Components(true, -n, DecimalKind.ZERO), ""));
            tests.add(new ParseCase("payloadzero " + n, "NaN" + zeros, new Components(false, DecimalKind.QNAN), ""));
            tests.add(new ParseCase("exponentzero " + n, "-0e-" + zeros, new Components(true, 0, DecimalKind.ZERO), ""));
            tests.add(new ParseCase("oversize coefficient " + n, "9".repeat(n), n <= 34 ? normal("9".repeat(n), 0) : null, ""));
            tests.add(new ParseCase("oversize payload " + n, "NaN" + "9".repeat(n), null, ""));
            tests.add(new ParseCase("oversize exponent " + n, "1e" + "9".repeat(n), null, ""));
            tests.add(new ParseCase("long trim " + n, " ".repeat(n) + "1" + "\t".repeat(n), normal("1", 0), ""));
            for (String tail : new String[]{"x", "\u00e9", "\n0"}) {
                for (String prefix : new String[]{"NaN", "1e", "1.", ""}) {
                    tests.add(new ParseCase("malformed tail " + n, prefix + zeros + tail, null, ""));
                }
            }
        }
        return tests;
    }

    private static void consume(ParseCase test) {
        try {
            sink = BidCodec.fromString(test.input());
        } catch (IllegalArgumentException error) {
            sink = error;
        }
    }

    private static void allocations(List<ParseCase> tests) {
        check(ManagementFactory.getThreadMXBean() instanceof ThreadMXBean, "ThreadMXBean unavailable");
        ThreadMXBean bean = (ThreadMXBean) ManagementFactory.getThreadMXBean();
        check(bean.isThreadAllocatedMemorySupported(), "thread allocation accounting unsupported");
        bean.setThreadAllocatedMemoryEnabled(true);
        long thread = Thread.currentThread().getId();
        for (int i = 0; i < 4096; i++) for (ParseCase test : tests.subList(0, 4)) consume(test);
        long max = 0;
        for (ParseCase test : tests) {
            for (int i = 0; i < WARMUP; i++) consume(test);
            long peak = 0;
            for (int sample = 0; sample < 3; sample++) {
                long before = bean.getThreadAllocatedBytes(thread);
                for (int i = 0; i < REPETITIONS; i++) consume(test);
                long allocated = bean.getThreadAllocatedBytes(thread) - before;
                check(before >= 0 && allocated >= 0, "invalid allocation counter");
                check(allocated <= (long) ALLOCATION_BUDGET * REPETITIONS,
                        test.name() + ": allocation " + allocated / REPETITIONS + " bytes/parse exceeds " + ALLOCATION_BUDGET);
                peak = Math.max(peak, (allocated + REPETITIONS - 1) / REPETITIONS);
            }
            max = Math.max(max, peak);
            if (test.input().length() >= 61760) {
                System.out.println("ALLOCATION java case=" + test.name() + " input_chars=" + test.input().length() + " peak_batch_bytes_per_parse=" + peak);
            }
            parseCase(test);
        }
        System.out.println("ALLOCATION java JDK=" + Runtime.version() + " method=ThreadMXBean thread_allocated_bytes"
                + " warmup_per_case=" + WARMUP + " samples=3 repetitions=" + REPETITIONS
                + " budget_bytes_per_parse=" + ALLOCATION_BUDGET + " observed_max=" + max
                + " input_generation=excluded budget_basis=bounded_34_digit_buffer_BigInteger_Components_or_exception_with_4KiB_headroom");
    }

    public static void main(String[] args) {
        widthBoundaries();
        List<ParseCase> tests = parserCases();
        for (ParseCase test : tests) parseCase(test);
        allocations(tests);
        System.out.println("PARSER-RESOURCE language=java cases=" + cases);
    }
}
