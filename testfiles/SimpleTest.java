class SimpleTest {

  int value;
  public int value2;
  private int value3;

  Test(int value) {
    this.value = value;
    this.value2 = value + 1;
    this.value3 = value + 2;
  }

  public int getValue(int specified) {
    if (specified == 1) {
      return this.value;
    } else if (specified == 2) {
      return this.value2;
    } else {
      return this.value3;
    }
  }

  public static String hello() {
    return "Hello World!";
  }
}
