/*
 * TestInterface tests for the correct handling of interfaces, with normal
 * methods, modifiers, and default methods
 */
public interface TestInterface {
  // Normal method
  boolean isTrue();
  // Method with parameters
  float getAt(int index);
  // Default method
  default int getIndex() {
    return 0;
  }
  // Public method, name should be capitalized
  public void publicMethod();
  // Annotated method
  @Unimportant
  int doesSomething();
}
