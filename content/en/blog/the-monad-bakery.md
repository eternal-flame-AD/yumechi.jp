---
title: "The Monad Bakery - the 65536th Monad Introduction"
description: An Introduction to Monad using a bakery analogy
date: 2024-09-01T19:55:26-05:00
image: 
math: 
license: CC BY-SA 4.0
hidden: false
comments: true
categories: ["Technical", "Haskell"]
tags: ["Technical", "Haskell", "Functional"]
---

## Introduction

When you write your first Haskell program, immediately you would come across the type `IO a`, called "the IO monad".

```haskell
main :: IO ()
main = putStrLn "Hello, World!"
```

If you go lookup what the definition of a `Monad` is, you would see it is a subtype of `Applicative`, and `Applicative` is a subtype of `Functor`. Very overwhelming, right?

Let's start with a simple definition and refine it as we go along. A `Monad a` is simply a value that has something to do with `a` but not quite an `a` by itself (recall that in Haskell, every value is pure).

For example it may be:
- a value that is of type `a` but is wrapped in a context, maybe some states or some side effects. (the above "Hello, World!" program is an example of this)
- a value `a` but is entered by the user at runtime.

## The Monad Bakery

Now let's look at the IO monad. Suppose we are a baker at a bakery. baking is a complicated process and cannot be modeled as a pure function. So we would have something like this:

```haskell
getFlour :: IO Flour -- we may need to buy more flour, side effects!
getWater :: IO Water
getYeast :: IO Yeast
mixIngredients :: Flour -> Water -> Yeast -> Dough -- nothing goes wrong here
bake :: Dough -> IO Bread -- the oven may explode, side effects!

recipeForBread :: IO Bread
recipeForBread = do
  flour <- getFlour
  water <- getWater
  yeast <- getYeast
  bakedBread <- bake (mixIngredients flour water yeast)
  pure bakedBread
```

What I wanted to express is, `IO Bread` is simply a "recipe" for `Bread`: it is not by itself a `Bread`, but it is a process that can produce a `Bread`.

### Functors

A `Functor` is a typeclass that maps values of one type to another type (note Haskell use the mathematical definition of a functor, it's different from the OOP functor).

Now let's look at why a `Monad` is a subtype of `Functor`. After you bake a bread, you may want to display it to the customer on a plate, or pack it in a bag to sell. Let's say you have these functions:

```haskell
display :: Bread -> Plate Bread
pack :: Bread -> Bag Bread
```

Note that these take the `Bread` as a value but we only have a recipe for `Bread`! However we can turn a recipe for `Bread` into a recipe for `Plate Bread` or `Bag Bread`:

```haskell
makeBreadForCustomer :: IO (Plate Bread)
makeBreadForCustomer = do
  bread <- recipeForBread
  pure (display bread)

makeBreadForSale :: IO (Bag Bread)
makeBreadForSale = do
  bread <- recipeForBread
  pure (pack bread)
```

Now we have a recipe for `Plate Bread` and `Bag Bread` respectively. However, being a `Functor`, we can simplify the code by using the `fmap` function provided by the `Functor` typeclass:

```haskell
makeBreadForCustomer :: IO (Plate Bread)
makeBreadForCustomer = fmap display recipeForBread

makeBreadForSale :: IO (Bag Bread)
makeBreadForSale = fmap pack recipeForBread
```

The infix operator `<$>` is just a synonym for `fmap`, so the above code can be written as:

```haskell
makeBreadForCustomer :: IO (Plate Bread)
makeBreadForCustomer = display <$> recipeForBread

makeBreadForSale :: IO (Bag Bread)
makeBreadForSale = pack <$> recipeForBread
```

There are a few helper in the `Functor` typeclass that are useful for composing functors:

- `(<$)` is defined as `f <$ a = fmap (const f) a`, which replaces the value inside the functor with a constant value. The flipped version `($>)` is also provided.
- `void` is defined as `void a = () <$ a`, which replaces the value inside the functor with `()`.

### Applicatives

An `Applicative` is a typeclass that allows you to apply a function that is wrapped in a context to a value that is also wrapped in a context. It is a concept that is more general than `Functor`. It took me quite a while to understand this when I first learned Haskell, so let's look at the `mixIngredients` function again:

```haskell
mixIngredients :: Flour -> Water -> Yeast -> Dough
```

How do I turn a recipe for `Flour`, `Water`, and `Yeast` into a recipe for `Dough`? 

Using the `do` notation, we can write:

```haskell
recipeForDough :: IO Dough
recipeForDough = do
  flour <- getFlour
  water <- getWater
  yeast <- getYeast
  pure (mixIngredients flour water yeast)
```

However this is just syntactic sugar, under the hood it is just a composition of functions by using the `<*>` operator provided by the `Applicative` typeclass:

```haskell
recipeForDough :: IO Dough
recipeForDough = mixIngredients <$> getFlour <*> getWater <*> getYeast
```

Elegant, isn't it?
How does this work?

Recall two facts:

1. All Haskell functions are curried, so applying an `a` to an `a -> b -> c` function gives you a `b -> c` function. 
2. `(->)` is right-associative, so `a -> b -> c` is a subtype of `a -> d` where `d` is bound to be `b -> c`.

So the type is inferred as:

```haskell
(<$>) :: Functor f => (a -> b) -> f a -> f b -- original definition
(<$>) :: (Flour -> Water -> Yeast -> Dough) -> IO Flour -> IO (Water -> Yeast -> Dough)

(<*>) :: Applicative f => f (a -> b) -> f a -> f b -- original definition
(<*>) :: IO (Water -> Yeast -> Dough) -> IO Water -> IO (Yeast -> Dough)
(<*>) :: IO (Yeast -> Dough) -> IO Yeast -> IO Dough
```

The first step uses `fmap` (a.k.a. `(<$>)`) to make a `Flour -> Water -> Yeast -> Dough` function into a `IO (Water -> Yeast -> Dough)` function, then the `<*>` operator applies the `Water` and `Yeast` sequentially to get a `IO Dough`.

There are a few helper functions in the `Applicative` typeclass that are useful for composing applicatives:

- `pure` is defined by the implementation of the `Applicative` typeclass, and it is used to lift a value into a context, a.k.a. `a -> f a`.
- `(<**>)`, which is a flipped version of `<*>`.'
- `liftA2` is defined as `liftA2 f a b = f <$> a <*> b`, which applies a binary function to two values in a context, similarly `liftA3` is also provided.
- `(<*)` and `(*>)` does two things in sequence but only returns the result of the first or second computation respectively.

### Monads

A `Monad` is simply an `Applicative` that allows you to "bind" the result of a computation to a function that returns a computation. Now that you have got your recipe for a `Dough` and a recipe for baking a `Dough` into a `Bread`, how can you get a recipe for `Bread`? You bind the result of the first recipe to the second recipe!

```haskell
-- the type signature of bind is:
(>>=) :: Monad m => m a -> (a -> m b) -> m b
-- which can be specialized to in our case:
(>>=) :: IO Dough -> (Dough -> IO Bread) -> IO Bread
```

### Putting it all together

Going back to the initial example, now we see that the `do` notation is just a syntactic sugar to make it look like you are writing imperative code, and the unsugared version is:

```haskell
recipeForBread :: IO Bread
recipeForBread = mixIngredients <$> getFlour <*> getWater <*> getYeast >>= bake
```

Note the operator precedence: `<$>` and `<*>` have higher precedence than `>>=`, so the above code is equivalent to:

```haskell
recipeForBread :: IO Bread
recipeForBread = (mixIngredients <$> getFlour <*> getWater <*> getYeast) >>= bake
```

Now we see that everything `do` notation does can be done purely functionally, they are just syntactic sugar for composing functions in these ways:

- Mapping the result of a computation to another type: `fmap` or `<$>`
- Apply the result of a computation to a function partially applied to the result of another computation: `<*>`
- Binding the result of a computation to another computation: `>>=`

Note that I'm not saying `do` notation is bad, it is a very useful tool for writing code that involve a lot of composing monadic values in a natural and readable way. However, for simple tasks like the above, it is better to use the unsugared version, which is more terse and faster to understand for a peer Haskell programmer.
However, use `do` notation generously if you expect someone who is not familiar with Haskell to read your code.

The commonly-used helper functions in the `Monad` typeclass are (some of them are only available in the `Control.Monad` module, not the prelude):

- `return` is just `pure` but only for `Monad` (I recommend writing `pure` instead of `return` unless it is the last statement in a `do` notation, as it is more general and distinguishable from the `return` in other languages).
- `sequence` converts a list of monadic values into a monadic list of values. This is used to convert a list of computations to a single computation that returns a list of values.
  - `sequence_` is similar to `sequence` but discards the result.
- `replicateM` repeats a computation a number of times and `sequence` the results.
  - `replicateM_` is similar to `replicateM` but discards the result.
- `forever` repeats a computation forever.
- `when` and `unless` conditionally execute a computation, basically shortcut of `if cond then action else pure ()`.
- `mapM` takes a function that returns a monadic value and applies it to a list of values, then `sequence` the results.
  - `mapM_` is similar to `mapM` but discards the result.
- `forM` is similar to `mapM` but with the arguments flipped, similarly `forM_` is `mapM_` with the arguments flipped.
- `foldM` is similar to `foldl` but the folding function rturns a monadic value.
- `(>>)` is identical to `(*>)`, but only available in the `Monad` typeclass.

## Generalization of Monads

Now that we have seen how `IO` can be modeled as a `Monad`, let's see what else can be modeled as a `Monad`. s

### Parser Combinators

A common task suitable for Haskell is parsing. Parser combinators are a way to build a parser by combining smaller parsers. We will use the built-in `Text.ParserCombinators.ReadP` module as an example:

Let's say you just bought a new recipe book and you want to parse the recipe for a bread. The recipe is in the format of:

```
Take: flour; 500g
Take: water; 300ml
Take: yeast; 10g

Mix: flour, water, yeast

Bake: flour, water, yeast; 200C, 30min
```

We can define a parser for the recipe:

```haskell
import Data.Char
import Data.Maybe (catMaybes)
import Text.ParserCombinators.ReadP

parseVerb :: ReadP String
parseVerb = string "Take" <++ string "Mix" <++ string "Bake"

parseIngredient :: ReadP (String, String, String)
parseIngredient = do
    name <- many1 (satisfy isAlpha)
    string "; "
    amount <- many1 (satisfy isDigit)
    unit <- string "g" <++ string "ml"
    pure (name, amount, unit)

data RecipeItem
    = Take (String, String, String)
    | Mix [String]
    | Bake [String] (String, String)
    deriving (Show)

parseRecipeItem :: ReadP RecipeItem
parseRecipeItem = do
    verb <- parseVerb
    string ": "
    case verb of
        "Take" -> do
            ingredient <- parseIngredient
            pure (Take ingredient)
        "Mix" -> do
            ingredients <- sepBy1 (many1 (satisfy isAlpha)) (string ", ")
            pure (Mix ingredients)
        "Bake" -> do
            ingredients <- sepBy1 (many1 (satisfy isAlpha)) (string ", ")
            string "; "
            temperature <- many1 (satisfy isDigit)
            string "C, "
            duration <- many1 (satisfy isDigit)
            string "min"
            pure (Bake ingredients (temperature, duration))

parseRecipe :: ReadP [RecipeItem]
parseRecipe = catMaybes <$> many (recipeItemLine <++ emptyLine)
  where
    emptyLine = Nothing <$ many (satisfy isSpace) <* string "\n"
    recipeItemLine = Just <$> parseRecipeItem <* string "\n"

recipeText :: String
recipeText =
    "Take: flour; 500g\n\
    \Take: water; 300ml\n\
    \Take: yeast; 10g\n\
    \\n\
    \Mix: flour, water, yeast\n\
    \\n\
    \Bake: flour, water, yeast; 200C, 30min\n"

main :: IO ()
main = print $ readP_to_S (parseRecipe <* eof) recipeText
```

This prints:

```
[([Take ("flour","500","g"),Take ("water","300","ml"),Take ("yeast","10","g"),Mix ["flour","water","yeast"],Bake ["flour","water","yeast"] ("200","30")],"")]
```

Note that these `ReadP a`'s are technically values in Haskell but they represent a computation that either produces a value of type `a` or fails, given the context of the input string. This is why `ReadP` is a `Monad`.

### State

Another common task suitable for Haskell is stateful computation. We will use the built-in `Control.Monad.State` module as an example:

Let's say you want to keep track of the amount of flour, water, and yeast you have in your bakery and write a recipe that automatically buys more ingredients when you run out. You can define a stateful computation to keep track of the inventory:

```haskell
import Control.Monad.State

data Inventory = Inventory
    { flour :: Int
    , yeast :: Int
    }
    deriving (Show)

emptyInventory :: Inventory
emptyInventory = Inventory 0 0

type Bakery = State Inventory

buyFlour :: Bakery ()
buyFlour = modify (\inv -> inv{flour = flour inv + 5000})

getFlour :: Int -> Bakery ()
getFlour amount = do
    curInv <- flour <$> get
    if curInv >= amount
        then modify (\inv -> inv{flour = curInv - amount})
        else buyFlour >> getFlour amount
```

### Monad Transformers

Here comes a problem: what if I need to bake a bread and I need both the `IO` and the `State` monad? You can use a _monad transformer_ to combine two monads into one. For example, you can use `StateT` to combine `State` and `IO`, putting it all together:

```haskell
import Control.Monad.State
import System.IO (hFlush, stdout)

data Flour = Flour
data Water = Water
data Yeast = Yeast

data Bread = Bread

bake :: Flour -> Yeast -> IO Bread
bake _ _ = putStrLn "Baking bread" >> pure Bread

data Inventory = Inventory
    { flour :: Int
    , yeast :: Int
    }
    deriving (Show)

emptyInventory :: Inventory
emptyInventory = Inventory 0 0

type BakeryT m = StateT Inventory m

buyFlour :: BakeryT IO ()
buyFlour = modify (\inv -> inv{flour = flour inv + 5000}) >> liftIO (putStrLn "Bought flour")

buyYeast :: BakeryT IO ()
buyYeast = modify (\inv -> inv{yeast = yeast inv + 500}) >> liftIO (putStrLn "Bought yeast")

getFlour :: Int -> BakeryT IO Flour
getFlour amount = do
    curInv <- flour <$> get
    if curInv > amount
        then Flour <$ modify (\inv -> inv{flour = curInv - amount})
        else buyFlour >> getFlour amount

getYeast :: Int -> BakeryT IO Yeast
getYeast amount = do
    curInv <- yeast <$> get
    if curInv > amount
        then Yeast <$ modify (\inv -> inv{yeast = curInv - amount})
        else buyYeast >> getYeast amount

bakeBread :: BakeryT IO Bread
bakeBread = do
    flour <- getFlour 500
    yeast <- getYeast 10
    liftIO $ bake flour yeast

runCommand :: BakeryT IO ()
runCommand = do
    liftIO $ putStr "bakery> " >> hFlush stdout
    input <- liftIO getLine
    if input == "exit"
        then pure ()
        else
            ( case input of
                "bake" -> () <$ bakeBread
                "inv" -> get >>= liftIO . print
                _ -> liftIO $ putStrLn "Invalid command"
            )
                >> runCommand

main :: IO ()
main = evalStateT runCommand emptyInventory
```

Now when we invoke `bakeBread`, we can be sure that we have enough ingredients to bake the bread!

```
bakery> inv
Inventory {flour = 0, yeast = 0}
bakery> bake
Bought flour
Bought yeast
Baking bread
bakery> inv 
Inventory {flour = 4500, yeast = 490}
bakery> bake
Baking bread
bakery> inv
Inventory {flour = 4000, yeast = 480}
bakery> bake
Baking bread
bakery> inv
Inventory {flour = 3500, yeast = 470}
```

Some more type wizardry: if you suddenly decide some functions only need to check the inventory but not use IO, you can use the `Identity` monad, which is a monad that does nothing but store a value:

```haskell
import Control.Monad.Identity

checkInventory :: BakeryT Identity Inventory
checkInventory = get
```

When you want to compose this into any other `BakeryT m` monad, you can use this `uplift` function:

```haskell
uplift :: (Monad m) => StateT s Identity a -> StateT s m a
uplift = StateT . (pure .) . runState
```

`runState` returns the new state and the value, `(pure .)` lifts the value into the `m` monad, and `StateT` lifts the function into the `StateT` monad transformer, neat!

## Extras

### Shortcuts to IO monad

Sometimes you don't really care about how an IO value is composed into your big `main :: IO ()`, and you just want to use the result lazily as a value. You can use the `unsafePerformIO` function to do this, but be careful as it can break referential transparency, as there is no guarantee that when the value is used you get a memoized value back or a new computation is done.

Note that it is not a very good idea to use `unsafePerformIO` as a kind of "lazy global variable" in your program, as it may be ran multiple times, instead compute it first and compose it with the rest of your program.

However one use case I use a lot in my study is to use it to emit intermediate steps in a computation, for example:

```haskell
{-# LANGUAGE RankNTypes #-}

type EmitFunc = forall a. a -> (a -> String) -> a

hEmitStep :: Maybe Handle -> EmitFunc
hEmitStep h a f = case h of
  Nothing -> a
  Just h' -> unsafePerformIO (hPutStrLn h' $ " --> " ++ f a) `seq` a
```

Recall that `seq` is a function that forces the evaluation of the first argument before returning the second argument, _when the second argument is needed_.

Here, we get a function that, when `a` is evaluated, writes out an intermediate step to a file handle. This is very useful when you are writing lazy algorithms and you want to see the intermediate steps that only involve computations that are actually done, and in order of when they are needed! In short, it produces a step-by-step trace of the computation that looks like a human would write it.

Let's see it in action! (note I changed `hEmitStep` a bit by `seq`ing the result before putting it to `hPutStrLn` to avoid the `--> ` from being printed before the actual value is computed (which may be a problem if the actual value itself prints more intermediate steps)):

```haskell
{-# LANGUAGE RankNTypes #-}

import Control.Monad (replicateM_)
import System.IO (Handle, hPutStrLn, openFile, stderr)
import System.IO.Unsafe (unsafePerformIO)

primes :: EmitFunc -> [Int]
primes emitStep = sieve [2 ..]
  where
    sieve (p : xs) =
        p `emitStep` (const $ show p ++ " must be prime!")
            : sieve
                [ x `emitStep` (const $ show x ++ " is not a multiple of " ++ show p)
                | x <- xs
                , (x `mod` p)
                    `emitStep` (\rem -> show x ++ " % " ++ show p ++ " = " ++ show rem)
                    /= 0
                ]

type EmitFunc = forall a. a -> (a -> String) -> a

hEmitStep :: Maybe Handle -> EmitFunc
hEmitStep h a f = case h of
    Nothing -> a
    Just h' -> unsafePerformIO (hPutStrLn h' $ f a `seq` "--> " ++ f a) `seq` a

main :: IO ()
main = do
    mapM_ (\prime -> putStrLn $ prime `seq` "Found prime: " ++ show prime) $
        take 5 $
            primes (hEmitStep (Just stderr))

```

This produces:

```
--> 2 must be prime!
Found prime: 2
--> 3 % 2 = 1
--> 3 is not a multiple of 2
--> 3 must be prime!
Found prime: 3
--> 4 % 2 = 0
--> 5 % 2 = 1
--> 5 is not a multiple of 2
--> 5 % 3 = 2
--> 5 is not a multiple of 3
--> 5 must be prime!
Found prime: 5
--> 6 % 2 = 0
--> 7 % 2 = 1
--> 7 is not a multiple of 2
--> 7 % 3 = 1
--> 7 is not a multiple of 3
--> 7 % 5 = 2
--> 7 is not a multiple of 5
--> 7 must be prime!
Found prime: 7
--> 8 % 2 = 0
--> 9 % 2 = 1
--> 9 is not a multiple of 2
--> 9 % 3 = 0
--> 10 % 2 = 0
--> 11 % 2 = 1
--> 11 is not a multiple of 2
--> 11 % 3 = 2
--> 11 is not a multiple of 3
--> 11 % 5 = 1
--> 11 is not a multiple of 5
--> 11 % 7 = 4
--> 11 is not a multiple of 7
--> 11 must be prime!
Found prime: 11
```

Nice! I only said to emit step when this value is needed, I never need to explicitly say when to emit the step, lazy functional programming does it for me! Additionally, if you use the values multiple times, it won't spam the output with the same step, as the value is memoized (in a reasonable way of course).

The same code would be very annoying to write in an imperative language, as you would need to explicitly write out the steps and the conditions to emit the steps, not to mention simulating infinite lists with iterators or callbacks. Additionally, an imperative implementation that does not involve emitting steps would look very different from the one that does, inside out. However here we only need to delete the `emitStep` calls and the program still works, just like you would write it without emitting steps: if you ask ChatGPT to write a prime program, it would look exactly like this, but without the `emitStep` calls!

```
write a terse infinite list of primes in Haskell with the sieve of eratoth  
algorithm, prefer conciseness over performance please.                     
                                                                            
Sure, here is a simple implementation:                                      
                                                                            
primes = sieve [2..]                                                      
    where sieve (p:xs) = p : sieve [x|x <- xs, x `mod` p > 0]

```

If you ask it to do it in Python, you get this:

```
write a terse infinite generator of primes in Python with the sieve of      
eratoth algorithm, perfer conciseness over performance please.              
                                                                            
Here is a concise Python generator function that implements the Sieve of    
Eratoth algorithm to generate an infinite sequence of prime numbers:        
                                                                            
def primes():                                                             
    D = {}                                                                
    q = 2                                                                 
    while True:                                                           
        if q not in D:                                                    
            yield q                                                       
            D[q*q] = [q]                                                  
        else:                                                             
            for p in D[q]:                                                
                D.setdefault(p+q, []).append(p)                           
            del D[q]                                                      
            q += 1 
```

It is really hard to make this code emit the correct steps, you might think that you can just add a `print()` after `D.setdefault()`, but no, you emit numbers before the previous prime numbers are emitted. This would print:

```
2
3
Eliminating 6, factors are [2]
5
Eliminating 8, factors are [2]
7
Eliminating 10, factors are [2]
Eliminating 12, factors are [3]
Eliminating 12, factors are [3, 2]
11
```

## Conclusion

In general, Haskell models non-pure computations as compositions of monads, and the program entry point is modeled as a single non-pure monad `IO ()`. This way, everything that seems to be impure in Haskell code is simply transforming small impure computations into larger impure computations, _in a pure way_.