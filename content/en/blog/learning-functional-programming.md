---
title: "Learning Functional Programming"
description: A brief delve into the presentation and performance of functional programs.
date: 2024-02-23T19:06:23-06:00
image: 
math: true
categories: ["Technical"]
tags: ["Technical", "Tutorial", "Haskell", "Rust", "Functional Programming"]
comments: true
---

## Introduction

I have been gradually learning Haskell for the past few months and I
have been fascinated by how different it is from the imperative
languages I am used to.

Here I document an experiment (which is based on one of my homework
problems, I just thought it is a great example) that I did where we
delve into comparing the presentation and the performance of code on a
spectrum from imperative to functional.

## Programming Challenge

Given the following theorem:

$$\begin{array}{r}
\text{ Given integers }a,b \geq 1\text{ : } \\
\sum_{r_{1} + \ldots + r_{a} = b}\binom{b}{r_{1}}\ldots\binom{b}{r_{a}} = \binom{ab}{b}
\end{array}$$

Write a function that takes two integers a and b as input and evaluates
both sides of the equation. Then use the function to verify that for
every $A \in \lbrack 1,10\rbrack$ and $B \in \lbrack 1,10\rbrack$, the
equation holds.

## Definitions

First we will provide an explanation of what it means to be imperative
and functional. (Disclaimer: this is my understanding of the concepts
and may not be 100% accurate)

-   In imperative languages, you write a sequence of statements that
    change the state of the program, and use control flow elements like
    loops, conditionals, function calls and gotos to control the flow of
    the program. Essentially it is a sequence of explicit instructions
    executed in a explicit order.

-   In (pure) functional languages, you do not write sequence of
    statements nor do you modify the state of the program. Instead you
    build up your program by composing functions and values together to
    form new functions and values.

    As a consequence you do not have control flow elements like loops,
    conditionals, function calls and gotos. Instead you use recursion,
    higher order functions (functions that transform other functions)
    and pattern matching to achieve the same effect.

## Actual Code Samples

We will provide three samples that gradually changes from an imperative
approach to a functional approach.

### Pure Imperative Approach

This is a nothing-special imperative approach to solve this problem.
Notice the frequent use of loops, conditionals and mutable states.

<div>

``` rs
use rug::Integer;
use std::time::Instant;

fn factorial(input: u64) -> Integer {
    let mut acc = 1.into();

    for i in 1..=input {
        acc *= i;
    }

    acc
}

fn multinomial_coef(k: &[u64]) -> Integer {
    let mut sum = 0;

    for &n in k {
        sum += n;
    }

    let nominator = factorial(sum);

    let mut denominator = Integer::from(1);
    for &n in k {
        denominator *= factorial(n);
    }

    nominator / denominator
}

fn binomial_coef(n: u64, k: u64) -> Integer {
    multinomial_coef(&[n - k, k])
}

pub fn sim(a: u64, b: u64) -> (Integer, Integer) {
    fn enumerate_ras<F: FnMut(Vec<u64>)>(a: u64, b: u64, result: &mut F) {
        fn enumerate_ras_inner<F: FnMut(Vec<u64>)>(
            rem_items: u64,
            rem_sum: u64,
            result: &mut F,
            current: Vec<u64>,
        ) {
            if rem_items == 0 && rem_sum == 0 {
                result(current);
                return;
            }
            if rem_items == 0 {
                return;
            }
            for next in 0..=rem_sum {
                let mut new_current = current.clone();
                new_current.push(next);
                enumerate_ras_inner(rem_items - 1, rem_sum - next, result, new_current);
            }
        }
        enumerate_ras_inner(a, b, result, Vec::new());
    }

    let lhs_single_item = |ras: Vec<u64>| {
        let mut acc = Integer::from(1);
        for r in ras {
            acc *= binomial_coef(b, r);
        }
        acc
    };

    let mut lhs = Integer::new();
    enumerate_ras(a, b, &mut |ras| {
        lhs += lhs_single_item(ras);
    });

    let rhs = binomial_coef(a * b, b);

    (lhs, rhs)
}

pub fn verify() -> (bool, usize) {
    let begin = Instant::now();
    for a in 1..10 {
        for b in 1..10 {
            let (lhs, rhs) = sim(a, b);
            if !(lhs == rhs) {
                return (false, begin.elapsed().as_millis() as usize);
            }
        }
    }
    (true, begin.elapsed().as_millis() as usize)
}
```

</div>

### Mixed Approach

A lot of modern languages are multi-paradigm, meaning they have both
facilities for imperative and functional programming. Rust is one of
them. In this example, we use a mix of imperative and functional
programming but substituting array iterations with higher order
functions.

<div>

``` rs
use rug::Integer;
use std::time::Instant;

fn factorial(input: u64) -> Integer {
    (1..=input).product()
}

fn multinomial_coef(k: &[u64]) -> Integer {
    factorial(k.iter().sum::<u64>())
        / k.iter()
            .map(|&n| factorial(n))
            .fold(Integer::from(1), |acc, n| acc * n)
}

fn binomial_coef(n: u64, k: u64) -> Integer {
    multinomial_coef(&[n - k, k])
}

pub fn sim(a: u64, b: u64) -> (Integer, Integer) {
    fn enumerate_ras<F: FnMut(Vec<u64>)>(a: u64, b: u64, result: &mut F) {
        fn enumerate_ras_inner<F: FnMut(Vec<u64>)>(
            rem_items: u64,
            rem_sum: u64,
            result: &mut F,
            current: Vec<u64>,
        ) {
            if rem_items == 0 && rem_sum == 0 {
                result(current);
                return;
            }
            if rem_items == 0 {
                return;
            }
            for next in 0..=rem_sum {
                let mut new_current = current.clone();
                new_current.push(next);
                enumerate_ras_inner(rem_items - 1, rem_sum - next, result, new_current);
            }
        }
        enumerate_ras_inner(a, b, result, Vec::new());
    }

    let lhs_single_item = |ras: Vec<u64>| {
        ras.iter()
            .map(|r| binomial_coef(b, *r))
            .fold(Integer::from(1), |acc, n| acc * n)
    };

    let mut lhs = Integer::new();
    enumerate_ras(a, b, &mut |ras| {
        lhs += lhs_single_item(ras);
    });

    let rhs = binomial_coef(a * b, b);

    (lhs, rhs)
}

pub fn verify() -> (bool, usize) {
    let begin = Instant::now();
    for a in 1..10 {
        for b in 1..10 {
            let (lhs, rhs) = sim(a, b);
            if !(lhs == rhs) {
                return (false, begin.elapsed().as_millis() as usize);
            }
        }
    }
    (true, begin.elapsed().as_millis() as usize)
}
```

</div>

### Pure Functional Approach

We write this program in Haskell, a purely functional programming
language. The only non-functional part of the program is computing the
CPU time. I wrote comments to assist in understanding the code and the
features of functional programming used in the code.

<div>

<div>

``` hs
import System.CPUTime (getCPUTime)

-- there is already a product function built in
-- that computes the product of a list of numbers
-- we only need to feed it the sequence [1..n]
-- no need to write any explicit loops or recursion
factorial :: (Integral n) => n -> n
factorial n = product [1 .. n]

mCoef :: (Integral n) => [n] -> n
-- to apply a function to each element of a list
-- we simply `map` the function over the list
-- no need to write any explicit loops or recursion
mCoef xs = factorial (sum xs) `div` product (map factorial xs)

-- this is a 'curried' function
-- which means it takes its arguments one at a time and returns a function
-- that takes the next argument, and so on
-- this is useful for 'partial application' where we supply some of the arguments
-- ahead of time and then supply the rest later, as we will use it below
bCoef :: (Integral n) => n -> n -> n
bCoef n k = mCoef [k, n - k]

sim :: Int -> Int -> (Integer, Integer)
sim a b =
  -- we can define some helper functions inside a function
  let
    -- here we define a function that takes a and b
    -- and generates all possible combinations of
    -- {r_1 ... r_a} such that r_1 + ... + r_a = b
    -- we notice that r_1 ... r_a can be chosen
    -- recursively where we choose r_1 first and then
    -- choose r_2 from the remaining sum, and so on
    -- all until we have chosen a times.
    makeRas :: Int -> Int -> [[Int]]
    -- this is called 'pattern matching'
    -- where the first pattern that matches the input is used
    -- for example the next line states that if a and b are both 0
    -- then return a list containing an empty list
    -- (because there is one way to choose 0 items from 0 items)
    makeRas 0 0 = [[]]
    -- this line states that if a is 0 (and b is not 0 is implied)
    -- then return an empty list (because there is no way to choose
    -- any items from 0 items)
    makeRas 0 _ = []
    -- then we have a catch-all recursive case
    -- that choosing {r_i ... r_a} is equivalent to choosing r_i
    -- first and then choosing {r_{i+1} ... r_a} from the remaining sum
    makeRas remItems remSum =
      -- this is called a 'list comprehension'
      -- it is a way to generate a list by iterating over multiple lists
      -- here we iterate over all possible choices of r_i as 'x'
      -- and then combine it with all possible choices of {r_{i+1} ... r_a}
      [x : xs | x <- [0 .. remSum], xs <- makeRas (remItems - 1) (remSum - x)]

    -- this computes a single summation term on the left hand side
    -- given a list of r values
    lhsSingleItem :: [Int] -> Integer
    -- this is a bit complex and may be confusing if you are not familiar with functional programming
    -- firstly 'fromIntegral' is simply used to convert between machine and arbitrary precision integers
    -- notice here we used the 'partial application' of 'bCoef'
    -- we first supplied the first argument to be 'b', and then
    -- we pass the partially applied function to 'map' to compute
    -- [bCoef b r_1, bCoef b r_2, ...]
    -- then we use 'product' to compute the product of all these values
    lhsSingleItem ras = product $ map (bCoef $ fromIntegral b) (map fromIntegral ras)

    -- compute the summation by summing the mapped values of 'lhsSingleItem' over all possible 'ras'
    lhs = sum $ map lhsSingleItem $ makeRas a b
    -- right hand side is trivial to compute, we just give it a name for consistency
    rhs = bCoef (fromIntegral (a * b)) (fromIntegral b)
   in
    -- this function return the pair (lhs, rhs)
    (lhs, rhs)

main :: IO ()
-- the 'do' block would be the closest resemblance to imperative programming
-- it actually strings operations together sequentially, like statements in an imperative language
-- (side note: even the 'do' notation is technically still just a composition of functions, it is just syntactic sugar)
main = do
  begin <- getCPUTime
  putStr
    $ show
    -- all is a function that takes a list of boolean values and returns whether
    -- all of them are True
    $ all
      -- uncurry as its name suggests convert a curried function to a function
      -- that takes a single pair of arguments
      -- notice that operators in Haskell are functions too
      -- so we can use 'uncurry' on the equality operator to
      -- get back a function that checks whether a pair of values are equal
      (uncurry (==))
    -- we use list comprehension again to generate all pairs of numbers from 1 to 10
    $ map (uncurry sim) [(x, y) | x <- [1 .. 10], y <- [1 .. 10]]
  end <- getCPUTime
  putStrLn $ "," ++ show (fromIntegral (end - begin) / 10 ^ 9)
```

</div>

</div>

## Performance

We measured two metrics for each code sample:

-   The number of non-empty non-comment lines in the code.

-   The time it takes to finish the verification.

<div>

``` r
lines <- read_csv("lines.csv")
timings <- read_csv("timings.csv")

p1 <- lines %>%
  ggplot(aes(x = paradigm, y = lines, fill = paradigm)) +
  geom_col() +
  geom_text(aes(label = lines), vjust = -0.5) +
  scale_y_continuous(expand = expansion(mult = c(0, 0.1), add = c(0, 10))) +
  labs(title = "Lines of Code in Different Paradigms",
       x = "Paradigm",
       y = "Lines of Code")

p2 <- timings %>%
  ggplot(aes(x = paradigm, y = time_ms, fill = paradigm)) +
  geom_boxplot() +
  labs(title = "Timing of Different Paradigms",
       x = "Paradigm",
       y = "Time (ms)")
plot_grid(p1, p2, ncol = 1)
```

</div>

![](/img/20230223-rust-haskell-compare.svg)

It seems that haskell is significantly more concise but also
significantly slower than the rust solutions. It seems that Haskell code
are notoriously hard to optimize and maybe GHC applies fewer
optimizations than rustc.

