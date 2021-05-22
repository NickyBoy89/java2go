package keywords

// List from https://www.w3schools.com/java/java_modifiers.asp
var (
  AccessModifiers = []string{"private", "protected", "public"}
  NonAccessModifiers = []string{"final", "static", "abstract", "transient", "synchronized", "volatile"}
)

var (
  PrimitiveTypes = []string{"byte", "short", "int", "long", "float", "double", "boolean", "char", "error"} // Error is just a workaround
)

// abstract 	continue 	for 	new 	switch
// assert*** 	default 	goto* 	package 	synchronized
// boolean 	do 	if 	private 	this
// break 	double 	implements 	protected 	throw
// byte 	else 	import 	public 	throws
// case 	enum**** 	instanceof 	return 	transient
// catch 	extends 	int 	short 	try
// char 	final 	interface 	static 	void
// class 	finally 	long 	strictfp** 	volatile
// const* 	float 	native 	super 	while
