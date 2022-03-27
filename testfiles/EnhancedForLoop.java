/*
 * Tests for correct behavior around enhanced for loops
 */
class EnhancedForLoop {
  public static void main(String[] args) {
    String[] words = {"this", "should", "be", "iterated", "over"};

    for (String word : words) {
      System.out.println(word);
    }
  }
}
