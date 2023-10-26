---
title: "Making a Discord Bot With Javacord"
description: 
date: 2023-10-26T01:14:17-05:00
categories: ["Technical"]
tags: ["Technical", "Tutorial", "Java", "Discord"]
hidden: false
comments: true
draft: true
---

## Registering a Discord Bot

### Creating the Application

Go to the [Discord Developer Portal](https://discord.com/developers/applications) and create a new application. Give it a name (and maybe a pfp). 

### Creating the Bot and Getting the Token

Then, go to the bot tab and create a bot. Give it a name (and maybe a pfp). Then hit "Reset Token" to get a token for authenticating your bot, note this down to later use in your code. Then in "Privileged Gateway Intents", enable the "Message Content Intent".


## Prepping the Project

### Creating the Project

Go to IntelliJ IDEA and create a new project. Select "Gradle" for the build system. Choose Kotlin as Gradle DSL. Then under Advanced -> GroupId put in your domain in reverse order (like jp.yumechi), if you do not have one put `io.github.<your github username>`. Then put in a name for the project. Then hit create.

![Creating a new project](/img/20231026-intellij-fxtwitterbot-create.png)

### Installing the Dependencies

1. Go to the `build.gradle.kts` file then:
2. Add the "application" plugin to the `plugins` block:

```kotlin
plugins {
    id("java")
    application
}
```

3. Then add the following lines to the `dependencies` block:

```kotlin
dependencies {
implementation("org.javacord:javacord:3.8.0")
}
```

4. Then register your main class at the bottom: 

```kotlin
application {
    mainClass.set("io.github.<your github username>.<your project name>.Main")
}
```

5. At last hit the "Sync" button in the top right corner.

![Updating the Dependencies](/img/20231026-fxtwitterbot-dep.png)

## Write the Code

### Creating the Main Class

First create a package in `src/main/java` that corresponds to the main class you set in the last step. For example if you set `jp.yumechi.FxTwitterBot.FxTwitterBot`, you will create a package called `jp.yumechi.FxTwitterBot` and then create a class named `FxTwitterBot` in that package.

Then add the following code to the class:

```java
package jp.yumechi.FxTwitterBot;

import org.javacord.api.DiscordApi;
import org.javacord.api.DiscordApiBuilder;
import org.javacord.api.entity.intent.Intent;

public class FxTwitterBot {
    public static void main(String[] args) {
        // Log the bot in
        DiscordApi api = new DiscordApiBuilder()
                .setToken("<your super secret token>")
                .addIntents(Intent.MESSAGE_CONTENT)
                .login().join();

        System.out.println("You can invite the bot by using the following url: " + api.createBotInvite());

        // Add a listener which answers with "Pong!" if someone writes "!ping"
        api.addMessageCreateListener(event -> {
            if (event.getMessageContent().equalsIgnoreCase("!ping")) {
                event.getChannel().sendMessage("Pong!");
            }
        });
    }
}

```

## Running the Bot Locally

Then in terminal (Alt-F12) run `./gradlew.bat run` (Windows) or `./gradlew run` (Linux/Mac) to run the bot locally. The program should print out an invite link to add your bot to your server.

## Deploying the Bot

TBD