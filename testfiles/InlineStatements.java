class InlineStatements {
  public static void main(String[] args) {
    int number = 4;

    if (number % 2 == 1) {
      System.out.println("Number is odd!");
    }

    // Test for inline if statements
    if (number % 2 == 0) System.out.println("Number is even!");

    System.out.println("The number is: " + number);
  }
}
