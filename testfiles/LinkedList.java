public class LinkedList {
	private int size = 0;

	private ListNode head;
	private ListNode tail;

	LinkedList() {
		this.head = null;
		this.tail = null;
	}

	public void add(ListNode elem) {
		this.add(elem, this.size);
	}

	public void add(ListNode elem, int index) {
		this.size++;
		if (this.size == 0) {
			this.head = elem;
			this.tail = elem;
			return;
		} else if (index > this.size - 1) {
			throw new IndexOutOfBoundsException("Specified index is larger than length of LinkedList");
		}

		if (index == this.size) {
			this.tail.next = elem;
			elem.prev = this.tail;
			this.tail = elem;
			return;
		} else if (index == 0) {
			this.head.prev = elem;
			elem.next = this.head;
			this.head = elem;
			return;
		}

		ListNode curNode = this.head;
		for (int i = 0; i < index; i++) {
			curNode = curNode.next;
		}

		curNode.next.prev = elem;
		elem.next = curNode.next;
		curNode.next = elem;
		elem.prev = curNode;
	}

	public int GetSize() {
		return this.size;
	}

	public String toString() {
		String result = "";
		System.out.println(result);
		ListNode curNode = this.head;
		for (int i = 0; i < this.size; i++) {
			result += " " + curNode.value;
			curNode = curNode.next;
		}
		return result;
	}

	class ListNode {
		ListNode next = null;
		ListNode prev = null;
		int value;

		public ListNode(int val) {
			this.value = val;
		}
	}

	public static void main(String[] args) {
		LinkedList linkedList = new LinkedList();

		int[] toAdd = {1, 2, 3, 4, 5, 6, 7, 8, 9};
		for (int i = 0; i < toAdd.length; i++) {

			ListNode newNode = new ListNode(toAdd[i]);
			linkedList.add(new ListNode(toAdd[i]));
		}

		System.out.println(linkedList);
	}
}
