import java.util.HashMap;
import java.util.Map;

/*
 * Tests for proper type inference in an untyped object
 */
public class UntypedMap {

  // This map has no type information supplied with it
  Map fruits;

  // This method supplies both correct types for both the key and value of the map
  public int getCount(String key) {
    return (int) this.fruits.get(key);
  }

  public static void main(String[] args) {
    UntypedMap storage = new UntypedMap();
    storage.fruits = new HashMap();
    storage.fruits.put("Apple", 1);
    storage.fruits.put("Orange", 3);
    System.out.println(storage.getCount("Apple"));
  }
}
