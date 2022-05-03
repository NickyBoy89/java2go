public class ResolveTesting {

  // This should be changed
  Node temp;

  private class Node {
    int value;

    public Node(int value) {
      this.value = value;
    }
  }

  // The parameter of this should change
  public void incrementNode(Node target) {
    target.value++;
  }

  public int square(int x1, int x2) {
    return x1 * x2;
  }

  public Node add(Node n1, Node n2) {
    Node temp = null;
    temp = new Node(n1.value + n2.value);
    return temp;
  }

  // The parameter and return type should change
  public Node duplicateNode(Node target) {
    return new Node(target.value);
  }
}
