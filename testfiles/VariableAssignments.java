/*
 * This file tests for the correct handling of assignments, and most natably:
 * Assigning a variable to a value in a statement e.g `int variable = 1;`,
 * using a value as an expression e.g `int variable = this.value = 1;`,
 * and nested assignments `int variable =  val1 = val2 = val3 = 3`
 */
public class VariableAssignments {
  public static void main(String[] args) {
    // Create and assign a variable
    int testVar = 0;

    // A variable with some modifier
    final int immutableVar = 1234;

    // Declare a variable that Go will guess wrong on, because the default
    // assumption for a floating point value is a float32
    double incorrectGuess = 1.0;

    // Assign a variable
    int variable;

    // Assign the variable to a new value in a statement
    variable = 1;

    // Set temp to the result of assigning 2 to `variable`
    int temp = variable = 2;
    // The temp variable should be 2
    System.out.println("This should be 2");
    System.out.println(temp);

    int var1;
    int var2;
    int var3;
    int var4;

    System.out.println(var1 = var2 = var3 = var4 = 10);
  }
}
