---
title: "Why I chose Rust over C and C++"
description: A Late Comment on the Recent Rust Situation in Linux Kernel
date: 2024-09-03T16:00:26-05:00
image: 
math: 
license: CC BY-SA 4.0
hidden: false
comments: true
categories: ["Technical", "Rust"]
tags: ["Technical", "C", "C++", "Programming", "Rust", "Software Engineering"]
---

I gave it some thought and decided to write a blog about why I prefer Rust over C and C++. 

## Background

Wedson Almeida Filho, a software engineer working on the Rust for Linux project, recently announced that he would be stepping down from the project, due to persistent "non-technical" issues mainly coming from a subset of the Linux kernel community who are not happy with the use of Rust in the Linux kernel.

[News Post](https://lwn.net/Articles/987635/)

Linux Torvalds responded to his decision with empathy while providing understanding for resistance of the adoption.

[News Post](https://www.zdnet.com/article/linus-torvalds-talks-ai-rust-adoption-and-why-the-linux-kernel-is-the-only-thing-that-matters/)

The source of resistance mainly comes from three main points, ordered by my perception of their importance in terms of technicality:

- C can be made safe with good practices, Rust is not necessary.
- Rust has more overhead than C, and can never be faster than theoretically optimized C.
- Most of the Linux kernel is written in C, and the community is used to it.

## Regarding Safety

I think the most important thing to clarify is that: It is not that we don't trust experienced developers to know how to write safe code, but rather that we want safety to be marked in the type system so they _can focus on, arguably more spectacular other things_.

### Memory Safety

In C, you can write a function like this:

```c
const char* string_trim(const char* str) {
    while (isspace(*str)) {
        str++;
    }
    return str;
}
```

You can also write a function like this:

```c
const char* string_trim(const char* str) {
    unsigned int trim = 0;
    while (isspace(str[trim])) {
        trim++;
    }
    char* trimmed = malloc(strlen(str) - trim + 1);
    strcpy(trimmed, str + trim);
    return trimmed;
}
```

They produce the same function signature, but requires totally different memory management practices:

- In the first version, it is the caller's responsibility to make sure the original string is alive when the result is used.
- In the second version, it is the caller's responsibility to free the result when it is no longer needed.

If you forgot to do this or you do the wrong branch, you get undefined behavior.

In Rust, `String` is a type that owns its memory, `&str` is a type that borrows memory. If you declare a function `fn string_trim(str: &str) -> &str`, the compiler will make sure that the result can only be dereferenced as long as the input is valid, and you don't need to worry about freeing the memory.

### Thread Safety

Let's say we have this thing, which is perfectly safe in C:

```c
#include <stdatomic.h>

typedef struct Progress {
    atomic_uint _done;
    unsigned int total;
} Progress;

void progress_inc(Progress *progress) {
    atomic_fetch_add(&progress->_done, 1);
}

unsigned int progress_get(Progress *progress) {
    return atomic_load(&progress->_done);
}
```

Let's say we have this other thing:

```c
typedef struct Progress {
    unsigned int done;
    unsigned int total;
} Progress;

void progress_inc(Progress *progress) {
    progress->done++;
}

unsigned int progress_get(Progress *progress) {
    return progress->done;
}
```

Both versions are correct in C, but the second one has more narrow assumptions. However, note that their API signatures are exactly the same. 

In Rust, the second version would be implicitly marked as `!Sync` which would only allow it to be used in a single-threaded context.

### Mitigation in C

You might say, I can just add a comment or add more abstractions to mark the intention clear. However, even if we put in the extra time to do that, eventually you start playing the game of "telephone" where the original intention gets lost in more and more layers of abstraction:

```c
typedef struct FileCopier {
    Progress progress;
    /* ... maybe some buffers and file handles ... */
} FileCopier;

void file_copier_get_progress(FileCopier *file_copier) {
    return progress_get(&file_copier->progress);
} 
```

Now, is this function thread-safe? It's not clear. It depends on the implementation of `FileCopier`, and the thread-safety of `FileCopier` depends on the implementation of `Progress`, see how it is hard to reason about even with just a single level of abstraction?

In Rust, you only need to write the following:

```rust
// Progress: Sync

struct FileCopier {
    progress: Progress,
    // ... maybe some buffers and file handles ...
}

impl FileCopier {
    fn get_progress(&self) -> &Progress {
        &self.progress
    }
}
```

Now, Rust automatically infers that while `get_progress()` itself depends on the safety of `FileCopier`, `Progress` is `Sync` and thus the result of `get_progress()` is `Sync`, which is the correct determination. 

If you only need to use `coper.progress` in a multi-threaded context, you should only pass `copier.get_progress()` to other threads instead of the whole `copier`.

## Regarding Performance

Rust has more overhead than C, which mainly comes from two sources:

- Bounds checking: Rust by default checks array bounds, which is not done in C.
- Memory management: Rust has more indirections in memory management which may lead to more unpredictable performance during drop.

### Bounds Checking

In Rust, you can bypass bounds checking by using `unsafe` functions:

```rust
fn main() {
    let mut array = [0; 1024];
    unsafe {
        for i in 0..1024 {
            *array.get_unchecked_mut(i) = i as i32;
        }
    }
}
```

If you really need the `[i]` syntax, you can use `std::ops::Index` and `std::ops::IndexMut` to implement it: (I know this is not technically conforming to Rust's safety rules, but it will not cause undefined behavior as long as you don't use it to access out-of-bound elements)

```rust
use std::ops::{Index, IndexMut, Deref, DerefMut};

#[repr(transparent)]
struct UnsafeVec<T>(pub Vec<T>);

// and similarly:
struct UnsafeSlice<'a, T>(pub &'a [T]); // or *const T

impl<T> Index<usize> for UnsafeVec<T> {
    type Output = T;

    fn index(&self, index: usize) -> &T {
        unsafe { self.0.get_unchecked(index) }
    }
}

impl<T> IndexMut<usize> for UnsafeVec<T> {
    fn index_mut(&mut self, index: usize) -> &mut T {
        unsafe { self.0.get_unchecked_mut(index) }
    }
}

impl Deref for UnsafeVec<T> {
    type Target = Vec<T>;

    fn deref(&self) -> &Vec<T> {
        &self.0
    }
}

impl DerefMut for UnsafeVec<T> {
    fn deref_mut(&mut self) -> &mut Vec<T> {
        &mut self.0
    }
}
```

### Memory Management

In Rust, you can use `unsafe` functions to manually manage memory, let's say we have this struct:

```c
struct DeviceInfo {
    char* name;
    char* vendor;
    char* model;
    char* serial;
    /* ... many more ... */
};
```

In high performance C, you would not allocate memory for each field separately, but rather allocate a single block of memory for the whole struct:

```c
DeviceInfo *device_info_new(const char *name, const char *vendor, const char *model, const char *serial) {
    DeviceInfo *device_info = malloc(sizeof(DeviceInfo));
    size_t total_length = strlen(name) + strlen(vendor) + strlen(model) + strlen(serial) + 4;

    char* block = malloc(total_length);

    device_info->name = block;
    strcpy(device_info->name, name);

    device_info->vendor = device_info->name + strlen(name) + 1;
    strcpy(device_info->vendor, vendor);

    device_info->model = device_info->vendor + strlen(vendor) + 1;
    strcpy(device_info->model, model);

    device_info->serial = device_info->model + strlen(model) + 1;
    strcpy(device_info->serial, serial);

    return device_info;
}

void device_info_free(DeviceInfo *device_info) {
    free(device_info->name);
    free(device_info);
}
```

However in Rust, if you declare the struct with `String`s, the compiler will force you to allocate and free each field separately:

```rust
struct DeviceInfo {
    name: String,
    vendor: String,
    model: String,
    serial: String,
    /* ... many more ... */
}

// implicit Drop implementation:
impl Drop for DeviceInfo {
    fn drop(&mut self) {
        drop(self.name);
        drop(self.vendor);
        drop(self.model);
        drop(self.serial);
    }
}
```

There are two solutions to circumvent this without runtime overhead nor exposes unsafe APIs:

Use self-referential structs:

```rust
#[ouroboros::self_referencing]
struct DeviceInfo {
    buf: String,
    #[borrows(buf)]
    name: &'this str,
    #[borrows(buf)]
    vendor: &'this str,
    #[borrows(buf)]
    model: &'this str,
    #[borrows(buf)]
    serial: &'this str,
}

macro_rules! slice_str {
    ($buf:expr, $start:expr, $end:expr) => {
        unsafe { $buf.get_unchecked($start..$end) }
    };
}

impl DeviceInfo {
    pub fn from_strs(name: &str, vendor: &str, model: &str, serial: &str) -> Self {
        // just to be compatible with C
        let buf = format!("{}\0{}\0{}\0{}\0", name, vendor, model, serial);
        DeviceInfoBuilder {
            buf,
            name_builder: |buf| slice_str!(buf, 0, name.len()),
            vendor_builder: |buf| slice_str!(buf, name.len() + 1, name.len() + 1 + vendor.len()),
            model_builder: |buf| slice_str!(buf, name.len() + 1 + vendor.len() + 1, name.len() + 1 + vendor.len() + 1 + model.len()),
            serial_builder: |buf| slice_str!(buf, name.len() + 1 + vendor.len() + 1 + model.len() + 1, name.len() + 1 + vendor.len() + 1 + model.len() + 1 + serial.len()),
        }.build()
    }
}
```

Use pointers:

```rust
// just to be compatible with C
struct DeviceInfo {
    buf: Vec<u8>,
    name: *mut u8,
    vendor: *mut u8,
    model: *mut u8,
    serial: *mut u8,
    /* ... many more ... */
}

impl DeviceInfo {
    pub fn from_strs(name: &str, vendor: &str, model: &str, serial: &str) -> Self {
        let mut buf = Vec::with_capacity(name.len() + vendor.len() + model.len() + serial.len() + 4);
        buf.extend_from_slice(name.as_bytes());
        buf.push(0);
        buf.extend_from_slice(vendor.as_bytes());
        buf.push(0);
        buf.extend_from_slice(model.as_bytes());
        buf.push(0);
        buf.extend_from_slice(serial.as_bytes());
        buf.push(0);

        let buf_ptr = buf.as_mut_ptr();
        let name_ptr = buf_ptr;
        let vendor_ptr = unsafe { buf_ptr.add(name.len() + 1) };
        let model_ptr = unsafe { vendor_ptr.add(vendor.len() + 1) };
        let serial_ptr = unsafe { model_ptr.add(model.len() + 1) };

        Self {
            buf,
            name: name_ptr,
            vendor: vendor_ptr,
            model: model_ptr,
            serial: serial_ptr,
        }
    }

    pub fn get_name(&self) -> &str {
        unsafe { std::ffi::CStr::from_ptr(self.name).to_str().unwrap() }
    }

    pub fn get_vendor(&self) -> &str {
        unsafe { std::ffi::CStr::from_ptr(self.vendor).to_str().unwrap() }
    }

    pub fn get_model(&self) -> &str {
        unsafe { std::ffi::CStr::from_ptr(self.model).to_str().unwrap() }
    }

    pub fn get_serial(&self) -> &str {
        unsafe { std::ffi::CStr::from_ptr(self.serial).to_str().unwrap() }
    }
}
```

## Regarding Community

I have not too much to say about this, which unfortunately is probably the most important blocker in the whole thing. One thing that is Rust's problem is that it is not as stable and standardized as C: too many times I need to do something and I found that I need to use nightly features to get what I want easily.

## Conclusion

Rust is a safe-by-default language, but it doesn't mean that it will handcuff you to do only what it can prove to be safe. It is a language that allows you to write unsafe code when you need to, but allows you to restrict unprovable safe code to a minimum scope so that when you are done with memory wizardry, you can focus more on your actual features.

Unfortunately the newness of Rust and the lack of standardization and stability is a big blocker for its adoption, but I think it should be a right direction to go for the future of software engineering. 