/*
 * This class tests for the correct conversion of the Java types to their Go
 * equivalents
 */
class RandomTypes {
  public static void main(String[] args) {
    /*
     * Some basic rules are as follows:
     * int -> int32
     * long -> int64
     * short -> int16
     * double -> float64
     * float -> float32
     * char -> rune
     * boolean -> bool
     * String - string
     */

    // Integral types
    int intType = 0;
    long longType = 0;
    short shortType = 0;

    // Floating point types
    float floatType = 0.0;
    double doubleType = 0.0;

    // Other random types:
    char character = 'a';
    boolean condition = false;

    // Specical cases
    String testString = "test";
  }
}
