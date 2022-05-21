package com.example;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

public class NameCollisions {

  Map<String, Integer> fruitTypes;

  // `map` is a reserved keyword
  Map<String, Integer> map;

  Map<String, Integer> test = new HashMap<>();

  // Since `type` is a reserved keyword in Go, this should fail
  public int getFruit(String type) {
    return this.fruitTypes.get(type);
  }

  // This is also a collision, with the keyword `range`
  private int[] range() {
    int[] values = new int[this.map.size()];
    int ind = 0;
    for (int val : this.map.values()) {
      values[ind] = val;
      ind++;
    }
    return values;
  }

  public NameCollisions() {
    // Another collision, but a little more subtle
    this.map = new HashMap<>();
  }

  public static void main(String[] args) {
    NameCollisions test = new NameCollisions();

    System.out.println(test.map);

    // Even more collisions
    Map<String, Integer> map = new HashMap<>();
    map.put("Apple", 1);
    map.put("Lemon", 4);

    test.map = map;
    System.out.println(Arrays.toString(test.range()));
  }
}
