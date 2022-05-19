class ScrambledForLoops {
  public static void main(String[] args) {
    // A completely normal for loop
    System.out.println("A completely normal for loop");
    for (int i = 0; i < 3; i++) {
      System.out.println(i);
    }

    // Remove the pre-condition statement from the loop. This should not change
    // how the loop behaves
    int j = 0;
    System.out.println("Remove the pre-condition");
    for (; j < 3; j++) {
      System.out.println(j);
    }

    // Remove the condition, this should not change behavior in this case
    System.out.println("Remove the statement");
    for (int i = 0; ; i++) {
      if (i >= 3) {
        break;
      }
      System.out.println(i);
    }

    // Remove the post-expression
    System.out.println("Remove the post-expression");
    for (int i = 0; i < 3; ) {
      System.out.println(i);
      i++;
    }

    System.out.println("Multiple declaration");
    int e;
    int f;
    for (e = 1, f = 1; e < 3; e++) {
      System.out.println(e);
    }
  }
}
