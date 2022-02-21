/*
 * This tests a specific case for `new` expressions, where they are specified
 * in a very specific way
 */

public class SelectorNewExpression {
  public static main(String[] args) {
    System.out.println("These should both be equal");
    // This tests the difference between calling a new constructor in two different ways
    // This test originally came from the following example calls:
    // Fernflower Output:
    EditGameRulesScreen.this.new RuleCategoryWidget();
    // CRF Output:
    new EditGameRulesScreen.RuleCategoryWidget();
  }
}
