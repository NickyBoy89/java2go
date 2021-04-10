import java.lang.AssertionError;

/*
 * Template for a singly- or doubly-linked list of ints.
 */
public class IntLinkedList {

	class Node {
		int data;
		Node next;

		Node(int data) {
			this.data = data;
			this.next = null;
		}
	}

	int size;
	Node head;
	Node tail;

	/**
	 * Construct an IntLinkedList.
	 */
	public IntLinkedList() {
		this.head = null;
		this.tail = null;
		this.size = 0;
	}

	/**
	 * Return the number of elements in the IntLinkedList.
	 *
	 * @return The number of elements in the IntLinkedList.
	 */
	public int size() {
		return this.size;
	}

	/**
	 * Add an element to the end of the IntLinkedList.
	 *
	 * @param element The element to add.
	 */
	public void add(int element) {
		if (this.head == null) { // Empty list
			this.size++;
			this.head = new Node(element);
			this.tail = this.head;
		} else {
			this.size++;
			this.tail.next = new Node(element);
			this.tail = this.tail.next;
		}
	}

	/**
	 * Get the element at the specified index.
	 *
	 * This function assumes that the index argument is within range of the IntLinkedList.
	 *
	 * @param index The index to get.
	 * @return The element at the specified index.
	 */
	public int get(int index) {
		Node currentNode = this.head;

		for (int i = 1; i <= index; i++) {
			currentNode = currentNode.next;
		}
		return currentNode.data;
	}

	/**
	 * Remove the element at the specified index.
	 *
	 * This function assumes that the index argument is within range of the IntLinkedList.
	 *
	 * @param index The index to remove.
	 */
	public void remove(int index) {
		Node currentNode = this.head;
		Node previousNode = this.head;

		if (index == 0) { // Special case for removing an element at the start of the list
			this.head = this.head.next;
		}

		for (int i = 1; i <= index; i++) {
			previousNode = currentNode;
			currentNode = previousNode.next;
		}

		if (currentNode.next == null) { // Special case for removing an element at the end of the list
			this.tail = previousNode;
		} else {
			previousNode.next = currentNode.next;
		}

		this.size--;
	}

	/**
	 * Create a String representation of the IntLinkedList.
	 *
	 * @return A String representation of the IntLinkedList.
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
	 * Check that an IntLinkedList contains the same elements as an int array.
	 *
	 * If the list and the array are not the same, throw an AssertionError.
	 *
	 * @param list The IntLinkedList to check.
	 * @param answer The expected answer, in the form of an int array.
	 */
	public static void assertArraysEqual(IntLinkedList list, int[] answer) {
		if (list.size() != answer.length) {
			throw new AssertionError("Expected list of length " + answer.length + " but got " + list.size());
		}
		for (int i = 0; i < answer.length; i++) {
			if (list.get(i) != answer[i]) {
				throw new AssertionError("Expected " + answer[i] + " but got " + list.get(i) + " at index " + i);
			}
		}
	}

	/*
	 * Test that the empty arraylist has size 0.
	 */
	public static void test1() {
		IntLinkedList list = new IntLinkedList();
		int[] answer = new int[0];
		assertArraysEqual(list, answer);
	}

	/*
	 * Test insertion into an arraylist (without resizing).
	 */
	public static void test2() {
		IntLinkedList list = new IntLinkedList();
		for (int i = 0; i < 3; i++) {
			list.add(i * i);
		}
		int[] answer = {0, 1, 4};
		assertArraysEqual(list, answer);
	}

	/*
	 * Test deletion from an arraylist without emptying it.
	 */
	public static void test3() {
		IntLinkedList list = new IntLinkedList();
		for (int i = 0; i < 5; i++) {
			list.add(i * i);
		}
		list.remove(1);
		list.remove(2);
		int[] answer = {0, 4, 16};
		IntLinkedList.assertArraysEqual(list, answer);
	}

	/*
	 * Test deletion from an arraylist and emptying it.
	 */
	public static void test4() {
		IntLinkedList list = new IntLinkedList();
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
		IntLinkedList.assertArraysEqual(list, answer1);

		// check that there are no last-element issues
		for (int i = 0; i < 5; i++) {
			list.add(i * i);
		}
		list.remove(4);
		list.add(-1);
		int[] answer2 = {0, 1, 4, 9, -1};
		IntLinkedList.assertArraysEqual(list, answer2);
	}

	/*
	 * Test insertion into an arraylist (with resizing).
	 */
	public static void test5() {
		IntLinkedList list = new IntLinkedList();
		for (int i = 0; i < 12; i++) {
			list.add(i * i);
		}
		int[] answer = {0, 1, 4, 9, 16, 25, 36, 49, 64, 81, 100, 121};
		IntLinkedList.assertArraysEqual(list, answer);
	}

	/**
	 * Put the IntLinkedList through some simple tests.
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
