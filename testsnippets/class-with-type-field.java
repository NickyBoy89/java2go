public class TestClass {
  private int value;
  // This right here is problematic, as "type" in Go
  // Happens to be a reserved keyword, and not usable in a struct field
  private String type;

  public int GetValue() {
    return this.value;
  }

  public String GetType() {
    return this.type;
  }

}
