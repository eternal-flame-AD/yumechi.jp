---
title: "The Monad Bakery - the 65536th Monad Introduction"
description: An Introduction to Monad using a bakery analogy
date: 2024-09-01T19:55:26-05:00
image: 
math: 
license: CC BY-SA 4.0
hidden: false
comments: true
categories: ["Technical"]
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

Note that I'm not saying `do` notation is bad, it is a very useful tool for writing code that involve a lot of composing monadic values in a natural and readable way. However, for simple tasks like the above, it is better to use the unsugared version, which is more terse and simpler. Additionally, use `do` notation generously if you expect someone who is not familiar with Haskell to read your code.

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
- `foldM` is similar to `foldl` but the folding function returns a monadic value.
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

type Bakery = State Inventory

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

## Conclusion

In general, Haskell models non-pure computations as compositions of monads, and the program entry point is modeled as a single non-pure monad `IO ()`. This way, everything that seems to be impure in Haskell code is simply transforming small impure computations into larger impure computations, _in a pure way_.