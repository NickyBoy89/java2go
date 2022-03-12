class ConditionOrdering {
  public static void main(String[] args) {
    int index = 10;

    // This tests a corner-case where the condition of the loop should
    // be parsed correctly

    for (int i = 0; i < index - 1; i++) {
      System.out.println(i);
    }
  }
}
