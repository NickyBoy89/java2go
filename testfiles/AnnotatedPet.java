public class AnnotatedPet {

  @Nullable private String name;

  public void setName(@Nullable String name) {
    this.name = name;
  }

  @Nullable
  public String getName() {
    return this.name;
  }

  @Deprecated
  @Environment(EnvType.CLIENT)
  public String sayHello() {
    return "Hello World!";
  }
}
