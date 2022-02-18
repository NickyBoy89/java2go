enum Compass {
	NORTH,
	SOUTH,
	EAST,
	WEST;

	public Compass() {}

	public static String hello() {
		return "Hello World";
	}

	public static void main(String[] args) {
		for (Compass c : Compass.values()) {
			System.out.println(c);
		}
	}
}
