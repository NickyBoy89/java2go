class MathUtils {

	public static final double PI = 3.14159;
	@Nullable
	public static final String NILSTRING = new String();

	public static int Min(int x1, int x2) {
		return x1 < x2 ? x1 : x2;
	}

	public static double GetPi() {
		return PI;
	}

	static {
		int i = 10;
		for (int j = 0; j < i; j++) {
			System.out.println(j);
		}
	}
}
