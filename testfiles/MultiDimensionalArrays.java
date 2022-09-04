import java.util.Arrays;

class MultiDimensionalArrays {
  public static void main(String[] args) {
    int[] test1 = new int[2];
    System.out.println(Arrays.toString(test1));

    int[][] test2 = new int[2][3];
    System.out.println(Arrays.deepToString(test2));

    int[][][] test3 = new int[2][3][4];
    System.out.println(Arrays.deepToString(test3));
  }
}
