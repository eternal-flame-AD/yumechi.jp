---
title: "Dynamically Load Assembler Code in R"
description: "An attempt at making R load and execute assembler code on the fly."
date: 2024-09-15T14:14:45-05:00
image: 
math: true
license: CC BY-SA 4.0
hidden: false
comments: true
categories: ["Technical", "R", "X86", "Abomination"]
tags: ["Technical", "R", "Assembler", "X86", "X86-64", "Rust", "Abomination"]
---

IMPORTANT: If the title isn't triggering enough, it should go without saying that this is a cool trick but terrible idea xD

The code is published along with the [typed-sexp](https://github.com/eternal-flame-AD/typed-sexp) library.

## Introduction

Imagine this frustration: You are not on your personal support computer, maybe a compute cluster or a friend is asking you 'why doesn't this work?'. You need a little native code to call a system API or do some heavy number crunching. You found that you don't have a working compiler that links with R! You can't even install a compiler because it's not your computer. And of course with this all traditional "inline" libraries won't work. How horrible! There has to be a way around this, right?

Introducing: `rasm`! A portable R extension that loads object files and executes them.

## Motivation

None.

Well, maybe one, malicious compliance: you see, CRAN has a policy that:

> Source packages may not contain any form of binary executable code.

It does't say anything about assembling ASM and loading it into the address space of the R process, right? :D

## Methodology

This section assumes you have a basic understanding of X86-64 assembly, memory paging and knows the memory representation of R objects as well as how R interacts with native code using the C FFIs. The last part can be learned from [Writing R Extensions](https://cran.r-project.org/doc/manuals/r-release/R-exts.html).

### Getting it to run

Firstly, let's make a short function that needs nothing but a stack and registers to run:

```nasm
section .text
    id: ; id <- function(x) x
        mov rax, rdi
        ret
```

This will take an R object and return it back.

This is not a difficult task for low-level developers, but for R users, I will provide an explanation on how this is done:

1. The assembler code need to be translated to machine code. This is done by an assembler, in this case NASM. This is a pure text-to-binary translation and there is no external dependency required. The resulting binary is called an _object file_, in Linux it has the `.o` extension and its format is called "ELF".
   
   The _object file_ is a container format that contains various "sections", each representing a different part of the program. In this case, the only section we defined is `.text`, which means executable code.

2. We need to put this code into memory so it can be executed. Memory is divided into _pages_, and each page has its own permissions. A page cannot be both writable and executable at the same time, so we need to request a new page `mmap()`, copy the code into it, and change the permissions to be executable `mprotect()`. Since we only have one section, we can just copy the `.text` section into the new page.
3. We need to be able to find this function when we want to call it. In this case, we need to parse the ELF file. In ELF, there is a `.symtab` section called the symbol table, which contains the names of the functions and their location in the file. We need to compute the actual address of the function at runtime by:
   $$
    \text{function address} = \text{page base address} + \text{offset}
   $$
4. We record all function offsets and their names in a table, and then move memory management for this table and the allocated page to R by using an "external pointer" (`EXTPTRSXP`) R object. This allows R to automatically clean up these resources when they are no longer needed.

### Making it useful

Just being able to run functions is not cool enough. I want to write assembly that can do anything a regular compiled shared library can do. Let's write a more complex example:

```nasm
section .rodata
    hello db "Hello, World!", 0

section .data
    counter dd 0

section .text
    extern Rf_ScalarInteger
    extern R_ShowMessage

    get_counter: ; SEXP(void)
        inc DWORD [counter]
        mov rdi, [counter]
        xor eax, eax
        call Rf_ScalarInteger
        ret
    
    hello_world: ; SEXP(SEXP)
        push rdi
        mov rdi, hello
        call R_ShowMessage
        pop rax
        ret
```

If we assemble this code and inspect it:

```asm
Disassembly of section .text:

0000000000000000 <get_counter>:
   0:   ff 04 25 00 00 00 00    inc    DWORD PTR ds:0x0
   7:   48 8b 3c 25 00 00 00    mov    rdi,QWORD PTR ds:0x0
   e:   00 
   f:   31 c0                   xor    eax,eax
  11:   e8 00 00 00 00          call   16 <get_counter+0x16>
  16:   c3                      ret

0000000000000017 <hello_world>:
  17:   57                      push   rdi
  18:   48 bf 00 00 00 00 00    movabs rdi,0x0
  1f:   00 00 00 
  22:   e8 00 00 00 00          call   27 <hello_world+0x10>
  27:   58                      pop    rax
  28:   c3                      ret
```

It's all zeros! But if you think about it, it makes sense. How will the assembler know the location of the counter relative to the `get_counter` function? It can't because it is the linker's job to put these sections together and we don't have a linker. So we need to do this manually.

The thing we need to do is called _relocation_. Which means adapting the code to it's actual loaded address by filling in these blanks the assembler left. We can list the relocations we need to do with `readelf -r`:

```bash
$ readelf -r get_counter.o
Relocation section '.rela.text' at offset 0x430 contains 5 entries:
  Offset          Info           Type           Sym. Value    Sym. Name + Addend
000000000003  00030000000b R_X86_64_32S      0000000000000000 .data + 0
00000000000b  00030000000b R_X86_64_32S      0000000000000000 .data + 0
000000000012  000900000002 R_X86_64_PC32     0000000000000000 Rf_ScalarInteger - 4
00000000001a  000200000001 R_X86_64_64       0000000000000000 .rodata + 0
000000000023  000a00000002 R_X86_64_PC32     0000000000000000 R_ShowMessage - 4
```

We see that we have 5 "blank" to fill in. So how do we do this?

Let define the problem more formally:

1. There is an incomplete instruction in the code which references to a location we only know at runtime.
2. We need to be able to figure out what exactly the relocation is asking for, resolve the address and patch the instruction so that it points to the correct location.

Let's try to read the first relocation, the important part is:

- Offset: Where do I want the relocation to be applied?
- Type: What kind of application do I want to do?
- Sym. Value: What is the *current* address of the relocation? (We haven't relocated yet so this is 0)
- Sym. Name: What is the name of the thing I want to reference?
- Addend: Where exactly do I want the address of, relative to the symbol?

Formally defined:

$$
\text{TransformFunc} :: \text{Ptr} \Rightarrow \text{Ptr} \Rightarrow \text{Ptr} \\
\text{transform} := \text{TransformOf}(\text{Type}) :: \text{TransformFunc} \\
\text{symValue} := \text{addrOf(referee(Sym. Name))} + \text{addend} \\
\text{dest} := \text{AddrOf(.text)} + \text{Offset} \\
\text{poke}(\text{dest} \lArr \text{transform}(\text{symValue}, \text{dest}))
$$

Now let's figure out how the transformation works:

We will look at the simpler of the two first, the `R_X86_64_64`: The name simply means "X86-64 platform, 64-bit absolute relocation". So the transformation is simply:

$$
\text{transform64Abs } \text{symValue } \text{dest} = \text{symValue}
$$

```rust
pub struct X6464Applicator {
    pub r_offset: u64,
    pub value: u64,
}

impl X6464Applicator {
    fn new(r_offset: u64, sym_val: usize, addend: i64) -> Self {
        let value = apply_addend(sym_val, addend) as u64;
        Self { r_offset, value }
    }
}

impl Applicator for X6464Applicator {
    fn apply_on_page(&self, dest: &Page) {
        unsafe {
            dest.as_ptr()
                .cast::<u64>()
                .byte_add(self.r_offset as usize)
                .write_unaligned(self.value);
        }
    }
}
```

The other one is `R_X86_64_PC32`: This means "X86-64 platform, 32-bit PC-relative relocation". This is a bit more complicated. The PC-relative means that the address is relative to the current instruction. So the transformation is:

$$
\text{transform32PC } \text{symValue } \text{dest} = \text{DWORD}(\text{symValue } - \text{dest})
$$

```rust
pub struct X64PC32Applicator {
    pub r_offset: u64,
    pub value: i64,
}

impl X64PC32Applicator {
    fn new(r_offset: u64, sym_val: usize, addend: i64) -> Self {
        let value = apply_addend(sym_val, addend) as i64;
        Self { r_offset, value }
    }
}

impl Applicator for X64PC32Applicator {
    fn apply_on_page(&self, dest: &Page) {
        let pc = dest.as_ptr() as i64 + self.r_offset as i64;
        unsafe {
            dest.as_ptr()
                .cast::<u32>()
                .byte_add(self.r_offset as usize)
                .write_unaligned(self.value.wrapping_sub(pc) as u32);
        }
    }
}
```

The last one is `R_X86_64_32S`: This means "X86-64 platform, 32-bit sign-extended relocation". It is the 32-bit version of the 64-bit absolute relocation. The transformation is:

$$
\text{transform32S } \text{symValue } \text{dest} = \text{DWORD}(\text{symValue})
$$

This one is a tricky one because we don't have control over how big $\text{symValue}$ is. We need to make sure that the sign extension is done correctly. This is done by checking whether sign-extension yield the same value as the original. If it doesn't, we will need to fail. There is a `nasm` warning option that can be used to make sure these instructions are not generated (`-Wreloc-abs-dword`).

```rust
impl X6432Applicator {
    fn new(r_offset: u64, sym_val: usize, addend: i64, sign_extend: bool) -> Self {
        let target_value = apply_addend(sym_val, addend);
        let sign32 = target_value & 0x8000_0000;
        let sign_extend_mask: usize = if sign_extend && sign32 != 0 {
            !0 >> 31 << 31
        } else {
            0
        };
        // We don't really have a say on where our page is mapped, so pointers can be really far
        assert_eq!(
            target_value,
            (target_value & 0xffff_ffff) | sign_extend_mask,
            "Relocation overflow, solution: use 64-bit instructions when accessing relocated data"
        );
        Self {
            r_offset,
            value: target_value as u32,
        }
    }
}

impl Applicator for X6432Applicator {
    fn apply_on_page(&self, dest: &Page) {
        unsafe {
            dest.as_ptr()
                .cast::<u32>()
                .byte_add(self.r_offset as usize)
                .write_unaligned(self.value);
        }
    }
}
```

The problem is some instructions in x86-64 does not take 64-bit address operands, so we need to write them in a different way to make it work. My solution is, for relocations within the assembly file, use relative addressing with `[rel my_data]` syntax. This way, the assembler will generate `R_X86_64_PC32` relocations, which we can make sure will fit. For external symbols, we will use 64-bit addressing by `movabs` into a temporary register first and then use the register from there.

Lastly, a short note about computing $\text{symValue}$, within the assembly file, the value is simply the loaded address of that section plus the offset of the symbol. For external symbols, we need to use the `dlsym` function to get the address of the symbol.

### Calling from R

We wrap everything we need from above into a struct:

```rust
#[derive(Debug)]
pub struct AsmFunction<I: ISA> {
    text: Page,
    #[allow(unused)]
    data: Option<Page>,
    #[allow(unused)]
    rodata: Option<Page>,
    /// Function offset table.
    func: HashMap<String, usize>,
    _pin: PhantomPinned,
    _isa: std::marker::PhantomData<I>,
}
```

We put this into a Box, and then into an [R external pointer](https://github.com/hadley/r-internals/blob/master/external-pointers.md) so that R can tell us when it's time to clean up.

```rust
#[export_name = "assemble"]
/// R external function to assemble a string into a module.
pub extern "C" fn assemble(input: SEXP) -> SEXP {
    let input = input
        .downcast_to::<CharacterVectorSEXP<_>>()
        .expect_r("input is not a string")
        .protect();

    if input.len() != 1 {
        Err::<(), _>("Expected a single string").unwrap_r();
    }
    let f = Box::new(
        AsmFunction::<X64ISA>::assemble(&input.get_elt(0).to_string())
            .expect_r("Failed to assemble"),
    );

    let ptr_inner = CharacterVectorSEXP::scalar("<asm_function>").protect();

    let ptr = Ptr::<SEXP, AsmFunction<X64ISA>>::wrap_boxed(f, r_nil(), ptr_inner);

    ptr.get_sexp()
}
```

Then, we do some macro magic to generate calling wrappers:

```rust
macro_rules! generate_asmcall {
    ($($name:ident( $( $arg_name:ident: $arg_ty:ident ),*))*) => {
        $(
            /// Call a function by name.
            #[cfg_attr(feature = "clobber_less", inline(never))] // let Rust clean up the registers as this function returns
            #[cfg_attr(not(feature = "clobber_less"), inline)]
            pub unsafe fn $name<R $(, $arg_ty)*>(&self, name: &str $(, $arg_name: $arg_ty)*) -> R {
                    let func =
                        self.text.as_ptr()
                            .byte_add(*self.func.get(name).expect("Function not found"));

                    let func = std::mem::transmute::<*const _, extern "C" fn($($arg_ty),*) -> R>(func);

                    log::debug!("Calling asm function {} at {:p}", name, func);

                    func($($arg_name),*)
            }
        )*
    }
}
```

## Demo

Let's see some demos in action!

### Glue Code

The library is really simple, just equivalent to this:

```R
dyn.load("librasm.so")

assemble <- function(asm, flavor = "nasm", isa = "x86_64") {
    if (flavor != "nasm") {
        stop("Only NASM is supported at the moment!")
    }
    if (isa != "x86_64") {
        stop("Only x86_64 is supported at the moment!")
    }
    .Call("assemble", asm)
}

.Asm <- function(box, name, ...) {
    .Call("asm_call", box, name, list(...))
}
```


### ForkR

For example, you want to call `fork` system call in Linux, you write this:

```R
wait <- function(pid) {
    exit_labels <- c("exited", "killed", "dumped", "trapped", "stopped", "continued")
    print(sprintf("I am the parent R process! My child is: %d", pid))
    status <- .Asm(asm, "waitpidr", pid) # Dummy, an exercise for the reader
    print(sprintf("My child exited with status: %s!", exit_labels[status]))
}

code <- file("forkR.s", "r")

asm <- assemble(paste(readLines(code), collapse = "\n"))

pid <- .Asm(asm, "forkr")

if (pid == 0) {
    print(sprintf("I am the child R process! My PID is: %d", Sys.getpid()))
    print("Crashing!")
    .Asm(asm, "crashpls")
}

wait(pid)
```

Output:

```
Forking a child process...
I'm the parent process from ASM!
I'm the child process from ASM!
[1] "I am the parent R process! My child is: 218008"
[1] "I am the child R process! My PID is: 218008"
[1] "Crashing!"
Crashing by executing UD2 in 3... 2... 1...

 *** caught illegal operation ***
address 0x7e8c0bf561ee, cause 'illegal operand'

Traceback:
 1: .Asm(asm, "crashpls")
An irrecoverable exception occurred. R is aborting now ...
[1] "My child exited with status: dumped!"
```

ASM code:

```nasm
.section .rodata
    notice_msg db "Forking a child process...", 0
    ; omitted for brevity

%define SIGINFO_T_SIZE 128
; more %define omitted for brevity

.section .text
forkr: ; SEXP(void)
    mov rdi, notice_msg
    call R_ShowMessage
    xor eax, eax

    mov rax, sys_fork
    syscall
    test rax, rax
    js .error
    push rax
    
    mov edi, eax
    call Rf_ScalarInteger
    mov rdi, rax
    call Rf_protect

    push rax

    mov rdi, success_parent_msg
    mov r8, success_child_msg
    mov rcx, [rsp + 8]
    test rcx, rcx
    cmovz rdi, r8
    mov rsi, rax
    call R_ShowMessage
    xor eax, eax

    mov rdi, 0x1
    call Rf_unprotect

    pop rax
    pop rcx
    ret

waitpidr: ; SEXP(SEXP)
    call INTEGER; now %rax is the pointer to the pid
    mov r12, [rax]
    mov rax, 0

    mov rdi, P_PID
    mov rsi, r12
    sub rsp, SIGINFO_T_SIZE
    mov rdx, rsp
    mov r10, WEXITED
    xor r8, r8

    mov rax, sys_waitid
    syscall
    test rax, rax
    js .error

    xor rdi, rdi
    mov edi, DWORD [rsp + SIGINFO_T_CODE_OFFSET]
    add rsp, SIGINFO_T_SIZE
    jmp Rf_ScalarInteger

crashpls: ; !(void)
    ud2
```

### SabotageR

You can also "sabotage" your R program by modifying the language itself :D

```R
code <- file("sabotageR.s", "r")

asm <- assemble(paste(readLines(code), collapse = "\n"))

y = 0 # I don't like this!!! Do something about it!

invisible(.Asm(asm, "sabotage", "="))

x <- 1
print(sprintf("`<-` still works! x is now: %d", x))
# [1] "`<-` still works! x is now: 1"

tryCatch({
    y = 2
}, error = function(e) {
    print(e) # <simpleError in y = 2: This is R, use <- instead of `=` :D
}, finally = {
    print(sprintf("y is still: %d", y)) # y is still: 0
})

# Error in y = 2 : This is R, use <- instead of `=` :D
# Execution halted
```

This works under the hood by modifying the built-in function table. However sky is the limit here: we are physically in a different page running machine code, we can just reprotect the R binary, patch it on the fly, protect it back and return to R.

Here's the ASM code:

```nasm

%define FUNTAB_SIZE 40
%define SEXPREC_HEADER_LEN 32

section .rodata
    crashpls_msg db "Something's seriously wrong, crashing by executing UD2 in 3... 2... 1...", 0
    do_set_not_found_msg db "do_set not found in the function table.", 10, 0
    no_eq_sign_msg db "This is R, use <- instead of `=` :D", 0
    equal_sign db "=", 0
    fmt_s db "%s", 0
    fmt_s_nl db "%s", 10, 0

section .data
    real_do_set dq 0
    eq_assign_call_no dq 0

section .text

    extern Rf_errorcall
    extern Rf_error
    extern Rf_ScalarInteger
    extern R_ShowMessage
    extern strcmp
    extern R_FunTab
    extern R_CHAR
    extern Rf_asChar

sabotage: ; SEXP(SEXP) 
    ; for the operator named in the parameter
    ; replace its entry in the function table with a custom wrapper
    call Rf_asChar
    mov rdi, rax
    call R_CHAR
    mov r14, rax
    mov r12, R_FunTab ; lea is too far away :(
    sub r12, FUNTAB_SIZE
    mov r13d, 0
    .loop:
        inc r13d
        add r12, FUNTAB_SIZE
        mov rax, [r12]
        test rax, rax
        jz .notfound
        mov rdi, rax 
        mov rsi, r14 
        call strcmp
        test rax, rax
        jnz .loop

        lea rax, [r12 + 8] ; get the function pointer
        mov rcx, real_do_set
        mov [rcx], rax ; save the original function pointer
        dec r13d
        mov rcx, eq_assign_call_no
        mov [rcx], r13d ; save the index

        mov r13, __patched_do_set
        mov DWORD [r12 + 8], r13d ; patch the table, evil >:)

        mov rdi, 0x1
        jmp Rf_ScalarInteger
    .notfound:
        xor rdi, rdi
        jmp Rf_ScalarInteger

__patched_do_set: ; SEXP(SEXP, SEXP, SEXP, SEXP) // rdi is the call, rsi is the discr
    push rdi
    push rsi
    push rdx
    push rcx

    xor rcx, rcx
    mov ecx, DWORD [rsi + SEXPREC_HEADER_LEN]
    mov r12, eq_assign_call_no ; just a demo, not really complete
    cmp rcx, [r12]
    je .is_equal_sign
    
    pop rcx
    pop rdx
    pop rsi
    pop rdi

    mov r12, real_do_set
    jmp [r12]
    jmp crashpls
    
    .is_equal_sign:
        push rcx
        xor eax, eax
        mov rsi, fmt_s_nl
        mov rdx, no_eq_sign_msg
        call Rf_errorcall
        jmp crashpls

crashpls:
    mov rdi, crashpls_msg
    call R_ShowMessage
    ud2
```

## Conclusion

This is a fun idea and I learned a lot about low-level programming and debugging. I guess it suits my interest since I was formerly an analyst in cybersecurity where we really do things as hacky as this and now I do data science, I am having some giggles on making R support inline assembly :D