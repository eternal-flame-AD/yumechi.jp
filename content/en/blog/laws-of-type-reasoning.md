---
title: "Laws of Type Reasoning"
description: A language-agnostic introduction to the laws of type reasoning in generic programming.
date: 2024-09-05T01:21:42-05:00
image: 
math: true
license: 
hidden: false
comments: true
draft: false
categories: ["Technical", "Type Theory"]
tags: ["Technical", "Rust", "C++", "Functional", "Generic Programming", "Type Theory"]
---

I'm not a type theorist, just someone who has been programming in various type systems for a while and wanted to write down some of the things I've learned. If you see any mistakes, please let me know!

## Introduction

Type reasoning is an important part of generic programming. It dictates what you can pass into and get out of a generic procedure. In this post, we will discuss the laws of type reasoning in generic programming and how to apply them in practice.

## Generic Refresher

A generic is simply a placeholder for a type yet to be determined. For example, in Rust, you can define a generic function like this:

```rust
fn id<T>(x: T) -> T {
    x
}
```

If you pass in an integer, you get an integer back. If you pass in a string slice, you get a string slice back. The type of `x` is determined by the caller, not the function itself.

## Quantification

Quantification is the process of determining the type of a generic. There are two types of quantification: universal and existential.

### Universal Quantification

Universal Quantification, called "for all" ($\forall$) in type theory, is the process of determining the type of a generic based on all possible types. For example, the type of the `id` function is universally quantified as follows:

$$
\textrm{id}: \forall T. T \to T
$$

Which means that for any type `T`, applying `id` to a value of type `T` will return a value of type `T`.

Here is a problem: If you want to explain a concept for someone, but you have no idea what they know, basically you can't explain anything. This is the same problem with universal quantification: it is not very useful to have a function that can take any type but has no way to interact with it.

Usually you will want to add some constraints to the type. For example, you might want to say that the type `T` must be able to be compared:

$$
\textrm{greater}: \forall T. (\textrm{Ord } T) \Rightarrow T \to T \to bool \\
$$

Now you know that although `T` can be anything, it must be a type that can be compared. This is called a constraint.

In Haskell, you write this as:

```haskell
greater :: Ord a => a -> a -> Bool
```

In Rust, you write this as:

```rust
fn greater<T: Ord>(x: T, y: T) -> bool { // or fn greater<T>(x: T, y: T) -> bool where T: Ord {
    x > y
}
```

C++ is an interesting case because until pretty recently there is actually no way to do this. You can't say that for a template with a type `T`, `T` must be comparable, the compiler can only just try to compile it can hope for the best.

### Existential Quantification

Existential Quantification, called "there exists" ($\exists$) in type theory, is the process of determining the type of a generic based on a specific type.
An example of existential quantification is a function that takes a value of an unknown type and returns a value of a wrapped type:

$$
\textrm{decrypt}: \forall T \exists D. T \to D \\
\textrm{eof}: \forall D. D \to bool \\
\textrm{runBlock}: \forall D. D \to string \\
$$

In this example, the `decrypt` function takes a value of type `T` and returns a value of type `D`. However, the type `D` is not decided by `T`, but by the implementation of `decrypt`. The `eof` function takes a value of type `D` and returns a boolean value. The `runBlock` function takes a value of type `D` and returns a string.

Here, we pass in a value we know the type of but get back a value whose type we don't know: we
only know that we can call `eof` and `runBlock` on it.

In Haskell, you write this as:

```haskell
{-# LANGUAGE ExistentialQuantification #-}

data D = forall a. Decryptor a

decrypt :: a -> D

eof :: D -> Bool
runBlock :: D -> String
```

Compiler implementations of existential quantification vary, but generally there are two options:

1. **Type Erasure**: Simply forget the actual type of `D` and only remembers how to do `eof` and `runBlock` on it. This implementation pattern is sometimes called dynamic dispatch. This way, there is only one version of `D`, `eof`, and `runBlock` for all `T`.
2. **Type Wrapping**: Wrap the actual type of `D` in a type that can be passed around. For example, `D` can be a `AesDecryptor<T>` concretely however the caller only knows it as `D` and thus can not use it as a real `AesDecryptor<T>`. This is called static dispatch: each `T` will need their own version of `D`, `eof`, and `runBlock`.

Rust supports both of these patterns:

```rust
struct DynamicD(Box<dyn Any>);

fn decrypt<T>(x: T) -> DynamicD {
    DynamicD(Box::new(x))
}

struct StaticD<T>(T);

trait Decryptor {
    fn eof(&self) -> bool;
    fn run_block(&self) -> String;
}

impl Decryptor for StaticD<AesDecryptor> {
    fn eof(&self) -> bool {
        todo!() 
    }

    fn run_block(&self) -> String {
        todo!()
    }
}

fn decrypt_static<T>(x: T) -> impl Decryptor { // exists D where D: Decryptor
    StaticD(x)
}
```

## Type Variance

Type variance is technically still a part of quantification, but it is important enough to warrant its own section.

### Covariance

When you say:

$$
\textrm{someValue}: \exists T. (\textrm{Ord } T) \Rightarrow T \\
\forall (\textrm{Ord } T): \textrm{Eq } T \\
\textrm{greater}: \forall T. (\textrm{Ord } T) \Rightarrow T \to T \to bool \\
\\
\textrm{equal}: \forall T. (\textrm{Eq } T) \Rightarrow T \to T \to bool \\
\\
$$

You are saying that if `T` is comparable, it is also equatable. This, even though we don't know the concrete type of `someValue`, we know that we can call `equal` on it because it is comparable.

### Contravariance

When you say:

$$
\forall (\textrm{Integral } T): \textrm{Num } T \\
\\
\textrm{findAnythingBigger}: \forall T \exists U. (\textrm{Num } T, \textrm{Num } U) \Rightarrow T \to U \\
$$

Does an `Integral` fit in `T`? Yes, because `Integral` is a subset of `Num`. Does a `U` fit in an `Integral`? No, because `U` is a superset of `Num`. Here we say `findAnythingBigger` is contravariant in `T` and covariant in `U`: a subtype of `Num` can be used at $(\textrm{Num } T)$, but a supertype of `Num` can be used at $(\textrm{Num } U)$.

### Invariance

Invariance is simply the combination of covariance and contravariance:

$$
\textrm{successor}: \forall T. (\textrm{Integral } T) \Rightarrow T \to T \\
$$

Here, `successor` is invariant in `T`: `T` must be an `Integral`, and the return value must also be an `Integral`, anything smaller or bigger will not work.

## RTTI

RTTI (Run-Time Type Information) is the process of determining the type of a value at runtime.
There are several important concepts related to RTTI:

### Reification

Reification is the process of turning a type into a value that you can compare with. For example, you may have this system:

$$
\textrm{intToString}: Int \to String \\
\textrm{boolToString}: Bool \to String \\
\\
\textrm{toString}: \forall T. T \to Option \langle String \rangle \\
$$

If you want to implement `toString` for all types, there is nothing you can do because T is not constrained in any way. However, if you reify the type:

$$
\textrm{intToString}: Int \to String \\
\textrm{boolToString}: Bool \to String \\
\\
\textrm{toString}: Int \to Option \langle String \rangle \\
\textrm{toString}: Bool \to Option \langle String \rangle \\
\textrm{toString}: \_ \to Option \langle String \rangle \\
$$

Now, the program will inspect the type of the value and call the appropriate `toString` implementation. 

### Introspection

Introspection is a system that allows you to actually inspect the type of a value at runtime. For example, you may have this system:

$$
\textrm{typeOf} : \forall T. T \to Type \\
\textrm{typeName} : Type \to String \\
\textrm{defaultValue} : \exists T. Type \to T \\
$$

This goes a step further than reification: you can actually ask the program to give you a value representation of a type that you can manipulate.

This is commonly used in dynamic languages like Python, and javascript, like this:

```typescript
function typeOf(x: any): string {
    return typeof x;
}
```

### Reflection

Reflection goes even further than introspection: it allows you to modify the value based on its type at runtime. For example, you may have this system:

$$
\textrm{Student} : Student \langle name: String, age: Int \rangle \\
\textrm{Employee} : Employee \langle name: String, salary: Int \rangle \\
$$

What if you want to just write a function that takes anything that has a `name` field and prints it? 

$$
\textrm{getName} : \forall T. T \to Option \langle String \rangle
$$

Neither the above two above options will work because you can't just ask for the value of the `name` field of a type `T`. However, with reflection, you can:

```go
func getName(x any) string {
    v := reflect.ValueOf(x)
    return v.FieldByName("name").String()
}
```

This is commonly available in languages that are statically typed but have a complex runtime, like Java, C#, and Go. The main application of this is serialization and deserialization, where you can simply declare a type and let the runtime will figure out how to make the data fit into it.

An additional use of reflection is runtime type inference, where you can take a value of an unknown type and determine whether the type satisfies a subtype at runtime.

```go
func asStudent(x any) *Student {
    if x, ok := x.(Student); ok {
        return &x
    }
    return nil
}
```

## Higher-Kinded Types

Higher-Kinded Types are types that are not quantified at a single point, for example:

$$
\textrm{AnyList}: \forall T. [ T ] \\
\textrm{ListAny}: [ \forall T. T ] \\
$$

A `[1, "hello"]` will fit into a `ListAny` but not into an `AnyList`: because `T` is quantified before the list is instantiated, the list can only contain one type, however in `ListAny`, `T` is quantified for each element of the list.

Functions can be higher-kinded as well, in fact it is the most common use of higher-kinded types:

$$
\textrm{PixelValueFunc}: \forall P. (\textrm{Pixel } P) \Rightarrow P \to Double \\
\textrm{avgIntensity}: \forall T. (\textrm{Image } T) \Rightarrow \textrm{ListAny } \langle T \rangle \to \textrm{PixelValueFunc} -> Double \\
$$

This takes a list of a mixture of any kind of image and a function that can convert any kind of image to a double, and returns the average of those values. In other words, I get a function that handles any image from a function that handles any pixel. The type of `P` is not determined by the caller of `avgIntensity`, but by the implementation of `avgIntensity` itself.

Not all statically typed languages support higher-kinded types, but some do, like Haskell and Scala:

```haskell
{-# LANGUAGE RankNTypes #-}

type AnyList = forall a. [a]
type ListAny = [forall a. a]
```

## Algebraic Data Types

This is a concept that is pretty much only used in TypeScript and Haskell, but it is interesting to mention. 

ADTs is the process of combining types to create new types by logical operations.

### Union Types

The simplest example is that if I know `T` is either a `string` or a `number`, then I can say that if `T` is not a `string`, it must be a `number`. 

This is useful because you might have this:

```typescript
type Phone = {
    type: "phone",
    number: string
}
type Email = {
    type: "email",
    address: string
}

type Contact = Phone | Email
```

Now, if you know that a `contact.type === "email"`, you can be sure that `contact.address` is a string.

### Intersection Types

The above is called union types, but there are also intersection types, which are types that are both of two types:

```typescript
type HasPhone = {
    phone: string
}

type HasEmail = {
    email: string
}

type HasAddress = {
    address: string
}

type OfficeEmployee = HasPhone & HasEmail & HasAddress
```

Now, if you have a `OfficeEmployee`, you know that it has a phone, email, and address.

### Negation and Exclusion

Additionally, there is type negation, which is the process of saying that a type is not a certain type:

```typescript
type RemoteEmployee = Exclude<EmployeeContact, HasAddress>
```

Now, if you have a `RemoteEmployee`, you know that it does not have an address.

There is also type exclusion, which is the process of saying that there can be no type that satisfies a certain condition:

```typescript
type HasNoAddress<T> = T & { address?: never }
```

This says that a `HasNoAddress<T>` is a `T` that can never have an `address` field, because `address` can only be `undefined` or `never`.

### Generalized Algebraic Data Types

GADTs are a generalization of the above concepts, where you can define types that are dependent on other types:

```typescript
type BiologyResearcher = {
    specialization: "genomics" | "developmental" | "ecology"
}

type PhysicsResearcher = {
    specialization: "quantum" | "relativity" | "particle"
}

type ResearchField = "biology" | "physics"

type Researcher<T extends ResearchField> = T extends "biology" ? BiologyResearcher : T extends "physics" ? PhysicsResearcher : never
```

Here, if `T` is `biology`, then `Researcher<T>` is `BiologyResearcher`, and if `T` is `physics`, then `Researcher<T>` is `PhysicsResearcher`, they are completely separate types but are dependent on the value of `T`.
