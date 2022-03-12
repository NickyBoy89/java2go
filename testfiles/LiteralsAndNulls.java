/*
 * This class tests
 */
class LiteralsAndNulls {
  public static void main(String[] args) {
    // Float types
    float n1 = 1.0F;
    double n2 = 1.0D;

    // Integer types
    long n3 = 1L;
    int n4 = 1;

    // Make sure objects are declared explicitly when they are `null`
    String n5 = null;

    System.out.println(n5);
  }
}
