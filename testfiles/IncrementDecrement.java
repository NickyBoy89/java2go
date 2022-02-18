/*
 * This tests for the correct handling of pre and post increment expressions,
 * as well as correct handling for them as statements
 */

public class IncrementDecrement {
  public static void main(String[] args) {
    // Since the increment for this loop is a statement, it should not be
    // generated as a pre-increment statement, and should be translated to a
    // native go `i++` statement
    for (int i = 0; i < 10; ++i) {
      System.out.println(String.format("The value of the varable is: %d", i));
    }

    int baseValue = 0;

    // This should be translated into a native go increment statement
    baseValue++;

    // This should be still translated into a native go increment statement
    ++baseValue;

    // Since the variable's value is `2`, and its value gets incremented after
    // it is evaluated, this should return `2`, and the variable should now
    // be `3`
    System.out.println("This should return 2");
    System.out.println(processValue(baseValue++));

    System.out.println("The value of the variable should be 3");
    System.out.println(baseValue);

    // The opposite, since `baseValue` is `3`, this increments the variable
    // before it is evaluated, this should return `4`, and the value should
    // be `4` as well
    System.out.println("This should return 4");
    System.out.println(processValue(++baseValue));

    System.out.println("The value of the variable should be 4 as well");
    System.out.println(baseValue);
  }

  public static int processValue(int value) {
    return value;
  }
}
