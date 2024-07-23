---
title: "Rust Functions (and Variance) in C++ Speak"
description: Rust functions, and variance, explained in C++ terms
date: 2024-07-23T10:07:17-05:00
image: 
math: 
license: CC BY-SA 4.0
hidden: false
comments: true
---

# Introduction

Rust functions contain more information than C++ functions. There are a total of 5 kinds of function-like things in Rust:

- Plain functions (`fn`)
- Function pointers (`*const fn`, `*mut fn`)
- Function traits:
    - `Fn`
    - `FnMut`
    - `FnOnce`

Let's break down each of these in C++ terms.

*All C++ code is compiled with `g++ -Wall -Wextra -pedantic -ggdb -std=c++17`.

# Definitions

## Plain Functions

Plain functions are the most straightforward. They are just like C++ functions. 

```rust
fn foo(x: i32) -> i32 {
    x + 1
}
```

is completely equivalent to (minus the overflow checks in debug mode):

```cpp
int foo(int x) {
    return x + 1;
}
```

Difference only begins when you start using references:

If we compile the following C++ code:

```cpp
char &simple_dangling()
{
    char c = 'E';
    return c;
}

char &find(std::string s, char c)
{
    for (char &x : s)
    {
        if (x == c)
        {
            return x;
        }
    }
    throw std::out_of_range("No such character found");
}

char &complex_dangling()
{
    std::string s;
    std::getline(std::cin, s);
    return find(s, 'E');
}
```

```
../../../test.cxx: In function ‘char& simple_dangling()’:
../../../test.cxx:8:12: warning: reference to local variable ‘c’ returned [-Wreturn-local-addr]
    8 |     return c;
      |            ^
../../../test.cxx:7:10: note: declared here
    7 |     char c = 'E';
      |          ^
```

GCC only catches the simple case of returning a reference to a local variable. It does not catch the more complex case of returning a reference to a local variable through a chain of function calls.

The same thing won't compile in Rust, in fact it won't even let you declare a function that returns a non-static reference without taking a reference as an argument.

```rust
fn dangling() -> &char {
    unreachable!()
}
```

```
 --> test.rs:1:18
  |
1 | fn dangling() -> &char {
  |                  ^ expected named lifetime parameter
  |
  = help: this function's return type contains a borrowed value, but there is no value for it to be borrowed from
help: consider using the `'static` lifetime
  |
1 | fn dangling() -> &'static char {
  |                   +++++++
```

When GCC sees the signature `char &dangling()`, it doesn't know that the returned reference is not a static reference so it will compile the function. Rust, on the other hand, knows that the reference is not static and will not compile the function unless you specify the reference refers to a static memory location.

We can get some clue on what rustc is doing by looking at the error message, it says that "this function's return type contains a borrowed value, but there is no value for it to be borrowed from". This means that all non-static references in plain functions must be borrowed from somewhere in the function's arguments, so if you change the signature to `fn not_dangling<'a>(_: &'a char) -> &'a char`, the function will compile. This `<'a>` is called a "lifetime parameter", it says that the reference returned by the function will be valid as long as the reference passed to the function is valid. This is similar to how generic works in C++, but instead of types, it's lifetimes: the compiler sees the calling code and works out the lifetime of the reference passed to the function, then it uses that lifetime to set the lifetime of the reference returned by the function.

If there is only one reference passed to the function, the lifetime parameter can be omitted and rustc will infer it, just write `fn not_dangling(_: &char) -> &char`.

## Function Pointers

Function pointers are the same as in C++, they are just pointers to functions. 

```rust
fn foo(x: i32) -> i32 {
    x + 1
}

fn main() {
    let f: fn(i32) -> i32 = foo;
    println!("{}", f(5));
}
```

The only difference is lifetimes are treated just like a type in C++, so you can't do this:

```rust
fn find(input: &str) -> Option<&str> {
    unimplemented!()
}

fn static_find(input: &str) -> Option<&'static str> {
    unimplemented!()
}

fn main() {
    let mut f = find;

    f = static_find;
}
```

## Function Traits

Function traits are the Rust version of C++'s functors. They are abstract struct types that overload the `()` operator. 

A plain function can be converted to a function trait:

```rust
fn find(input: &str) -> Option<&str> {
    unimplemented!()
}

fn main() {
    let f: &dyn Fn(&str) -> Option<&str> = &find;
}
```

Rust also have closures, just like C++ lambdas: they are just syntactic sugar for creating an anonymous struct that overloads the `()` operator.

```rust
fn main() {
    let f = |x: i32| x + 1;
    println!("{}", f(5));
}
```

However, things get more complicated when you start capturing variables outside the closure, and that's where the three subtraits of `Fn` come in:

### `Fn`

`Fn` is the most restrictive of the three in terms of what the closure can do. These closures only immutably borrow variables from the outside scope.

```rust
fn main() {
    let x = 5;
    let f: &dyn Fn() -> i32 = &|| x + 1;
    println!("{}", f()); // 6
}
```

Since this function cannot mutate any captured variables, it can be passed to and called anywhere, anytime, and as many times as you want.

### `FnMut`

`FnMut` is the middle ground. These closures can mutate variables from the outside scope.

```rust
fn main() {
    let mut x = 5;
    let mut f = |y| x += y;
    f(1);
    println!("{}", x); // 6
}
```

Note that `f` is declared as mutable, this restricts the shared use of `f`. This is because since `f` can mutate captured reference `x`, it is no longer safe to call `f` at the same time at multiple places: one caller needs to "finish" with `f` before another caller can use `f`.

`FnMut` is a superset of `Fn`.

### `FnOnce`

`FnOnce` is the most permissive of the three in terms of what the closure can do. These closures "consume" variables from the outside scope. It may free the memory of the captured variables, or destroy the captured variables in some other way.

```rust
fn main() {
    let x = Some("hello".to_string());
    let get: Box<dyn FnOnce() -> String> = Box::new(|| x.unwrap());
    println!("{}", get());
}
```

Here `get` takes the memory of the variable contained in `x`, so `x` is no longer usable after `get` is called, thus once this function is called, `x` is destroyed, and `get` is no longer usable because there is no `x` to take the memory from.

`FnOnce` is a superset of `FnMut`.

# On the Topic of Variance

## Covariance

Covariance is a simple concept: if `A` is a subtype of `B`, then `A` can be used wherever `B` is expected, just like how subclasses can be used wherever the superclass is expected in C++.

When we apply this to Rust lifetimes, let's look at the following code:

```rust
fn weird_find<'a>(mut input: &'a str, target: u8) -> Option<&'a u8> {
    let static_str: &'static str = "Hello, world!";

    input = static_str;

    input.as_bytes().iter().find(|&&x| x == target)
}

fn main() {
    let input = String::from("Goodbye, world!");
    let target = b'o';

    let result = weird_find(&input, target);

    println!("{:?}", result);
}
```

In `main()`, rustc infers that `'a` in `weird_find` is the same as the lifetime of `input`, a local variable. However in `weird_find`, `input` is reassigned to a static string, which has a `'static` lifetime. Since rustc knows that `'static` outlives any other lifetime, it allows the assignment. This is covariance in action: the `'a` in `weird_find` is a **supertype** of `'static`, so it can be assigned to a `'static` reference.

In C++, this can be demonstrated with the following code:

```cpp
struct SuperClass
{
};

struct SubClass : public SuperClass
{
};

void take_superclass(const SuperClass &super)
{
}

int main()
{
    const SubClass sub;

    take_superclass(sub);
}
```

## Contravariance

However this actually relationship get's inverted when we start looking at function traits. 
Let's look at the following code:

```rust
fn main() {
    let static_find = Box::new(|x: &'static str| -> Option<&'static u8> {
        let equals = |x: &u8| -> bool { *x == b'o' };

        for c in x.as_bytes() {
            if equals(c) {
                return Some(c);
            }
        }

        None
    });

    let s = "hello";

    let result = static_find(s);
}
```

This code compiles, rustc infers that since `static_find` gets a static reference, the reference must be valid for the lifetime of `equals`, in Rust speak we say that the `x` in `equals` is contravariant with respect to the `x` in `static_find`: the `x` in `static_find` is a **subtype** of the `x` in `equals`, so it can be passed to `equals`.

The same can be observed in C++ as well:

```cpp
struct SuperClass
{
};

struct SubClass : public SuperClass
{
};

void correct()
{
    const auto take_subclass = [](const SubClass &sub)
    {
        const auto take_superclass = [](const SuperClass &super) {
        };

        take_superclass(sub);
    };

    SubClass sub;

    take_subclass(sub);
}

void incorrect()
{
    const auto take_superclass = [](const SuperClass &super)
    {
        const auto take_subclass = [](const SubClass &sub) {
        };

        take_subclass(super);
    };

    SubClass sub;

    take_superclass(sub);
}
```

```
test.cxx: In lambda function:
test.cxx:31:22: error: no match for call to ‘(const incorrect()::<lambda(const SuperClass&)>::<lambda(const SubClass&)>) (const SuperClass&)’
   31 |         take_subclass(super);
      |         ~~~~~~~~~~~~~^~~~~~~
test.cxx:31:22: note: candidate: ‘void (*)(const SubClass&)’ (conversion)
test.cxx:31:22: note:   candidate expects 2 arguments, 2 provided
test.cxx:28:36: note: candidate: ‘incorrect()::<lambda(const SuperClass&)>::<lambda(const SubClass&)>’
   28 |         const auto take_subclass = [](const SubClass &sub) {
      |                                    ^
test.cxx:28:36: note:   no known conversion for argument 1 from ‘const SuperClass’ to ‘const SubClass&’
```

## Invariance

Invariance is the simplest concept: something that is invariant only accepts the exact type it expects, without considering subtypes or supertypes.
This is most commonly seen where a type needed to be both covariant and contravariant.

```rust
fn find<'a>(mut input: &'a str, target: u8) -> Option<&'a u8> {
    input.as_bytes().iter().find(|&&x| x == target)
}

fn main() {
    let input = String::from("Goodbye, world!");
    let target = b'o';

    let result = find(&input, target);

    println!("{:?}", result);
}
```

Here, the lifetime of `input` is `'a`, and the lifetime of the reference returned by `find` is also `'a`. This is invariance in action: the lifetime of the reference returned by `find` is the same as the lifetime of `input`, so the reference returned by `find` must be exactly the same lifetime as the reference passed to `find`.

In C++, this can be demonstrated with the following code:

```cpp
struct SuperClass
{
};

struct SubClass : public SuperClass
{
};

struct SubSubClass : public SubClass
{
};

void invariant()
{
    const auto take_superclass = [](const SuperClass &super)
    {
        return super;
    };

    const auto take_subclass = [&take_superclass](const SubClass &sub)
    {
        return static_cast<const SubClass &>(take_superclass(sub));
    };

    SuperClass super;
    super = take_subclass(super);

    SubClass sub;
    sub = take_subclass(sub);

    SubSubClass subsub;
    subsub = take_subclass(subsub);
}
```

# Conclusion

Rust has a lot of analogues to C++, but builds on top of them to create a more expressive and safer language. The lifetimes in Rust are a powerful tool that can be used to enforce memory safety and prevent reference errors. Function traits track memory side effects of closures, and variance allows the compiler to reason about the relationships between different types and lifetimes allowing for more flexible usage of functions and variables.