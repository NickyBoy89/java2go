public class GenericLinkedList<T> {
  int size;
  Node head;
  Node tail;

  class Node {
    T val;
    Node next;

    Node(T data) { // Node inherits T from GenericLinkedList
      this.val = data;
    }
  }

  /** Construct an GenericLinkedList. */
  public GenericLinkedList() {}

  /**
   * Return the number of elements in the GenericLinkedList.
   *
   * @return The number of elements in the GenericLinkedList.
   */
  public int size() {
    return this.size;
  }

  /**
   * Add an element to the end of the GenericLinkedList.
   *
   * @param element The element to add.
   */
  public void add(T element) {
    this.size++;
    Node newNode = new Node(element);
    if (this.head == null) {
      this.head = newNode;
      this.tail = newNode;
      return;
    }

    this.tail.next = newNode;
    this.tail = newNode;
  }

  /**
   * Get the element at the specified index.
   *
   * <p>This function assumes that the index argument is within range of the GenericLinkedList.
   *
   * @param index The index to get.
   * @return The element at the specified index.
   */
  public T get(int index) {
    Node curNode = this.head;
    for (int i = 0; i < this.size; i++) {
      if (i == index) {
        return curNode.val;
      }
      curNode = curNode.next;
    }
    return null;
  }

  /**
   * Remove the element at the specified index.
   *
   * <p>This function assumes that the index argument is within range of the GenericLinkedList.
   *
   * @param index The index to remove.
   */
  public void remove(int index) {
    Node curNode = this.head;
    this.size--;
    if (index == 0) {
      this.head = this.head.next;
      return;
    }

    for (int i = 0; i < index - 1; i++) {
      curNode = curNode.next;
    }

    // Tail
    if (curNode.next.next == null) {
      this.tail = curNode;
      curNode.next = null;
    } else {
      curNode.next = curNode.next.next;
    }
  }

  /**
   * Create a String representation of the GenericLinkedList.
   *
   * @return A String representation of the GenericLinkedList.
   */
  public String toString() {
    String result = "{";
    if (this.size() > 0) {
      result += this.get(0);
    }
    for (int i = 1; i < this.size; i++) {
      result += ", " + this.get(i);
    }
    result += "}";
    return result;
  }

  /**
   * Check that an GenericLinkedList contains the same elements as an int array.
   *
   * <p>If the list and the array are not the same, throw an AssertionError.
   *
   * @param list The GenericLinkedList to check.
   * @param answer The expected answer, in the form of an int array.
   */
  public static void assertArraysEqual(GenericLinkedList list, int[] answer) {
    if (list.size() != answer.length) {
      throw new AssertionError(
          "Expected list of length " + answer.length + " but got " + list.size());
    }
    for (int i = 0; i < answer.length; i++) {
      if ((Integer) list.get(i) != answer[i]) {
        throw new AssertionError(
            "Expected " + answer[i] + " but got " + list.get(i) + " at index " + i);
      }
    }
  }

  /*
   * Test that the empty arraylist has size 0.
   */
  public static void test1() {
    GenericLinkedList<Integer> list = new GenericLinkedList<>();
    int[] answer = new int[0];
    assertArraysEqual(list, answer);
  }

  /*
   * Test insertion into an arraylist (without resizing).
   */
  public static void test2() {
    GenericLinkedList<Integer> list = new GenericLinkedList<>();
    for (int i = 0; i < 3; i++) {
      list.add((Integer) i * i);
    }
    int[] answer = {0, 1, 4};
    assertArraysEqual(list, answer);
  }

  /*
   * Test deletion from an arraylist without emptying it.
   */
  public static void test3() {
    GenericLinkedList<Integer> list = new GenericLinkedList<>();
    for (int i = 0; i < 5; i++) {
      list.add(i * i);
    }
    list.remove(1);
    list.remove(2);
    int[] answer = {0, 4, 16};
    GenericLinkedList.assertArraysEqual(list, answer);
  }

  /*
   * Test deletion from an arraylist and emptying it.
   */
  public static void test4() {
    GenericLinkedList<Integer> list = new GenericLinkedList<>();
    for (int i = 0; i < 5; i++) {
      list.add(i * i);
    }

    list.remove(1);
    list.remove(2);

    // delete the final remaining numbers
    list.remove(0);
    list.remove(0);
    list.remove(0);
    int[] answer1 = {};
    GenericLinkedList.assertArraysEqual(list, answer1);

    // check that there are no last-element issues
    for (int i = 0; i < 5; i++) {
      list.add(i * i);
    }
    list.remove(4);
    list.add(-1);
    int[] answer2 = {0, 1, 4, 9, -1};
    GenericLinkedList.assertArraysEqual(list, answer2);
  }

  /*
   * Test insertion into an arraylist (with resizing).
   */
  public static void test5() {
    GenericLinkedList<Integer> list = new GenericLinkedList<>();
    for (int i = 0; i < 12; i++) {
      list.add(i * i);
    }
    int[] answer = {0, 1, 4, 9, 16, 25, 36, 49, 64, 81, 100, 121};
    GenericLinkedList.assertArraysEqual(list, answer);
  }

  /**
   * Put the GenericLinkedList through some simple tests.
   *
   * @param args Ignored command line arguments.
   */
  public static void main(String[] args) {
    test1();
    test2();
    test3();
    test4();
    test5();

    System.out.println("pass");
  }
}
