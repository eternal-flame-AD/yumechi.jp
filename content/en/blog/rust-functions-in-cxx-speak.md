---
title: "Rust Functions (and Variance) in C++ Speak"
description: Rust functions, and variance, explained in C++ terms
date: 2024-07-23T10:07:17-05:00
image: 
math: 
license: CC BY-SA 4.0
hidden: false
comments: true
categories: ["Technical", "Rust"]
tags: ["Technical", "Rust", "C++", "Functional"]
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

However, things get more complicated when you start capturing variables outside the closure, and Rust needs to know the kind of operations the closure will do to the captured variables, and that's where the three subtraits of `Fn` come in:

### `Fn`

`Fn` is the most restrictive of the three in terms of what the closure can do. These closures only immutably borrow variables from the outside scope.

```rust
fn main() {
    let x = 5;
    let f: &dyn Fn() -> i32 = &|| x + 1;
    println!("{}", x); // 5
    println!("{}", f()); // 6
}
```

Since this function cannot mutate any captured variables, it can be passed to and called anywhere, anytime, and as many times as you want (as long as the captured variables are still valid).

In C++, all lambdas are treated as `Fn` closures, so you can pass them around and call them as many times as you want without any restrictions.

Side note: If you want to capture a variable by value, you can use the `move` keyword, however this invalidates the original variable:

```rust
fn main() {
    let x = Box::new(5); // Use a heap-allocated variable so the value is not simply copied
    let f: &dyn Fn() -> i32 = &move || *x + 1;
    println!("{}", x); // Error
    println!("{}", f()); // 6
}
```
```
error[E0382]: borrow of moved value: `x`
 --> test.rs:4:20
  |
2 |     let x = Box::new(5);
  |         - move occurs because `x` has type `Box<i32>`, which does not implement the `Copy` trait
3 |     let f: &dyn Fn() -> i32 = &move || *x + 1;
  |                                ------- -- variable moved due to use in closure
  |                                |
  |                                value moved into closure here
4 |     println!("{}", x);
  |                    ^ value borrowed here after move
  |
  = note: this error originates in the macro `$crate::format_args_nl` which comes from the expansion of the macro `println` (in Nightly builds, run with -Z macro-backtrace for more info)

error: aborting due to 1 previous error

For more information about this error, try `rustc --explain E0382`.
```

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

The commonly demonstrated "counter" example is a good example of `FnMut`:

```rust
fn make_counter() -> impl FnMut() -> i32 {
    let mut count = 0;
    move || {
        count += 1;
        count
    }
}

fn main() {
    let mut counter = make_counter();
    println!("{}", counter()); // 1
    println!("{}", counter()); // 2
    println!("{}", counter()); // 3
}
```

Note that the `count` variable is declared as a local variable, but it becomes "owned" by the closure, so its memory is not freed when `make_counter` returns.

Under the hood, this closure actually gets "pulled out" from `make_counter` and `count` becomes part of the stack frame for `main`.

```asm
example::make_counter:
 mov    DWORD PTR [rsp-0x4],0x0
 mov    eax,DWORD PTR [rsp-0x4]
 ret
 nop    DWORD PTR [rax]

example::make_counter::{{closure}}:
 sub    rsp,0x18
 mov    QWORD PTR [rsp+0x8],rdi
 mov    eax,DWORD PTR [rdi]
 inc    eax
 <...>

main:

 <...>
 lea    rdi,[rsp+0xc]
 call   8720 <example::make_counter::{{closure}}>
 <...>
```

The same thing is not possible in C++, some correct and incorrect alternatives are:

```cpp
std::function<int()> invalid_make_counter()
{
    // incorrect
    int count = 0;
    return [&count]()
    { return count++; };
}

std::function<int()> smart_counter()
{
    // correct but wasteful
    auto count = std::make_shared<int>(0);
    return [count]()
    {
        return (*count)++;
    };
}

struct ManualCounter
{
    int count = 0;
    int operator()()
    {
        return count++;
    }
};

std::function<int()> make_counter()
{
    auto counter = ManualCounter{};
    return counter;
}

std::function<int()> leaky_counter()
{
    // incorrect
    int *count = new int{0};
    return [count]()
    {
        return (*count)++;
    };
}
```

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

Here `get` takes the memory of the variable contained in `x`, so `x` is no longer usable after `get` is called, thus once this function is called, `x` is destroyed, and consequently `get` also becomes invalid because there is no `x` to take the memory from.

C++ closures cannot be specified as "once-only". This makes double-free possible:

```cpp
std::function<void()> download_something()
{
    auto buf = new std::vector<char>();

    return [buf]()
    {
        std::copy("Hello, World!", "Hello, World!" + 13, std::back_inserter(*buf));
        std::cout << std::string(buf->begin(), buf->end()) << std::endl;
        delete buf;
    };
}

int main()
{
    auto doit = download_something();

    doit(); // "Hello, World!"
    doit(); // Address 0x4e4c080 is 0 bytes inside a block of size 24 free'd

    return 0;
}
```

`FnOnce` is a superset of `FnMut`.

# On the Topic of Variance

Variance is a relationship between types that describes how subtypes and supertypes can be used in place of each other. There are three kinds of variance:

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

However this relationship actually get's inverted when we start looking at function traits. 
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

This code compiles, rustc infers that since `static_find` gets a static reference, the reference must be valid for the lifetime of `equals`, in Rust speak we say that the lifetime parameter in `equals` is contravariant with respect to the lifetime parameter in `static_find`: the `x` in `static_find` is a **subtype** of the `x` in `equals`, so it can be passed to `equals`.

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
fn outer_find<'a>(x: &'a str) -> Option<&'a u8> {
    let correct_inner_find = |x: &'a str| -> Option<&'a u8> {
        for c in x.as_bytes() {
            if *c == b'h' {
                return Some(c);
            }
        }

        None
    };
    let too_strict_inner_find = |x: &str| -> Option<&u8> {
        for c in x.as_bytes() {
            if *c == b'h' {
                return Some(c);
            }
        }

        None
    };
    let too_lax_inner_find = |mut x: &'a str| -> Option<&'static u8> {
        x = "another string";
        for c in x.as_bytes() {
            if *c == b'h' {
                return Some(c);
            }
        }

        None
    };

    return correct_inner_find(x);
    return too_strict_inner_find(x);
    return too_lax_inner_find(x);
}

fn main() {
    let s = "hello";

    let result = outer_find(s);
}
```
```
error: lifetime may not live long enough
  --> test.rs:14:24
   |
11 |     let too_strict_inner_find = |x: &str| -> Option<&u8> {
   |                                     -               - let's call the lifetime of this reference `'2`
   |                                     |
   |                                     let's call the lifetime of this reference `'1`
...
14 |                 return Some(c);
   |                        ^^^^^^^ returning this value requires that `'1` must outlive `'2`

error: lifetime may not live long enough
  --> test.rs:24:24
   |
1  | fn outer_find<'a>(x: &'a str) -> Option<&'a u8> {
   |               -- lifetime `'a` defined here
...
24 |                 return Some(c);
   |                        ^^^^^^^ returning this value requires that `'a` must outlive `'static`
```

Here, any closure that returns a reference that is not exactly the same lifetime as the input reference will not compile. This is invariance in action: the `x` in `outer_fin` is both covariant and contravariant with respect to the `'a` lifetime, thus it can only return a reference that is exactly the same lifetime as the input reference.

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

int main()
{
    const auto invariant = [](const SubClass &subClass) -> const SubClass &
    {
        const auto correct = [](const SubClass &subClass) -> const SubClass &
        {
            return subClass;
        };
        const auto too_lax = [](const SuperClass &superClass) -> const SuperClass &
        {
            return superClass;
        };
        const auto too_strict = [](const SubSubClass &subSubClass) -> const SubSubClass &
        {
            return subSubClass;
        };

        return correct(subClass);
        return too_lax(subClass);
        return too_strict(subClass);
    };

    SubClass subClass;
    const auto &ret = invariant(subClass);

    return 0;
}
```
```
test.cxx: In lambda function:
test.cxx:33:23: error: invalid initialization of reference of type ‘const SubClass&’ from expression of type ‘const SuperClass’
   33 |         return too_lax(subClass);
      |                ~~~~~~~^~~~~~~~~~
test.cxx:34:26: error: no match for call to ‘(const main()::<lambda(const SubClass&)>::<lambda(const SubSubClass&)>) (const SubClass&)’
   34 |         return too_strict(subClass);
      |                ~~~~~~~~~~^~~~~~~~~~
```

You can manually `static_cast` the `SuperClass` reference to a `SubClass` reference, but it is an unchecked cast and the compiler won't do it for you.

# Conclusion

Rust has a lot of analogues to C++, but builds on top of them to create a more expressive and safer language. The lifetimes in Rust are a powerful tool that can be used to enforce memory safety and prevent reference errors. Function traits track memory side effects of closures, and variance allows the compiler to reason about the relationships between different types and lifetimes allowing for more flexible usage of functions and variables.