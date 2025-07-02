---
title: "uBO ruleset for \"First Party\" Behavioral Tracking and Data Exfiltration"
date: 2025-07-01T11:56:14-05:00
categories: ["Technical", "Privacy"]
tags: ["Technical", "Privacy", "Ongoing Work"]
---

## The evolution of Analytics

The line between sites collecting aggregated metrics and active spying by session recording and stitching together the "journey" of a particular user's every click and scroll from impression to conversion is becoming increasingly blurred. The latter form of data collection is extremely invasive as they collect granular, non-aggregated, and almost never properly anonymized data, which can be immediately used to "optimize" the persuasiveness (and ultimately the revenue) of the content by knowing exactly at which point the user is most likely to be persuaded.

While traditionally the frontier of the argument generally hinges on whether users visiting a specific site are "implicitly consenting" to some data points being tracked,
the introduction of technologies such as [session replay](https://posthog.com/session-replay) and [CDP session stitching](https://github.com/dbt-labs/segment/blob/main/models/sessionization/docs.md) are increasingly entering the realm of non-consensual tracking.

## What is "implicit consent" vs "behavioral" analytics?

For example, if I purchased a book from a bookstore, and halfway into the book I noticed on the jacket there is a sequel to the book, so I went to the same bookstore and purchased the sequel. In this scenario, I do believe there is some implicit consent in that the bookstore is agreeing on using this behavioral data to indicate this book is popular. If someone at the bookstore is being smart they might even notice the book is popular to my age and gender category. I don't see a problem with that.

However, what if the bookstore installed a "recorder" into the book and recorded the exact duration I read the book every night and cross reference it with the date I purchased more books so the author can "optimize" their future books to be more profitable by adding a big red button just at that point in the book to future customers? Would that still be considered "implicit consent" that comes with the transaction? Would I feel comfortable providing that data knowing that bookstores will use that data to psychologically persuade other people to purchase the sequel when they would otherwise not have purely based on **the value of the persuasive content alone**? Does that degrade the value of the book itself from a piece of literature and art to the amount of persuading power it has?

## The evolution of "first party" tracking

While traditional trackers like Google Analytics are considered "bad", "unethical", and "centralized", they are at least transparent about their intent and purpose and give you limited recourse to opt out (you can opt out to the use of your data against yourself but NOT against other users). They provide the shovel (how to collect data) and users some control over how personalized the data collection can be used "against you". However, a new generation of "privacy preserving" trackers and SSPs focus on legal compliance (such as GDPR requiring 'third party' processors to be compliant) however they find their marketing niche to "bypassing tracking blockers" by explicitly documenting and endorsing how to [**bypass "tracking blockers"**](https://plausible.io/docs/proxy/introduction) by using reverse proxies through CDNS (usually these methods are used for bypassing state sponsored censorship, thus typical CNAME cloaking detection will NOT be effective). 

I believe no reasonable person would agree that a vendor knowing what "tracking blockers" mean (users dissent to tracking) and then advocating to their clients to take a targeted, active step to bypass these dissenting users is in any shape or form a tracking practice based on "implicit consent" of the transaction (user visits your website).

It seems like some AdTech and SSPs are trying to exploit the loophole of "privacy" in legal compliance (which generally draws the line between "first party" and "third party" tracking) but completely sidestepping or even in some ways further exploiting users who explicitly dissent to tracking.

However, these companies seem to fail to understand basic principles of statistics in human study, or the fact that it is the user's computer, brain and wallet after all:

- It's the user's computer, so finding a way to bypass the blocker today because you (first-party tracking) are not considered a threat is not a technical "marvel", but rather an evolving threat landscape that need to be addressed (which this post is trying to do). When the users are forced to take aggressive action by inviting users to assess this risk from a cosmetic issue (the "clean web" argument) to **Security Threat of their own behavioral autonomy**, it reveals the end game: you contribute to a **less operational web environment** by ruining Canvas, ruining third-party cookies, ruining WebRTC, ruining WebGL, ruining WebSocket, and so on.
- It's the user's brain, in human studies "complete data" does not exist. No matter what study you do there will be participants who wish to opt out at any time, usually not at random. MNAR methods exist to account for this. It is generally also not considered unethical to estimate the cohort of users who opt out by observing edge metrics to make this a "known unknown". However, trying to game the system by collecting limited data from non-consenting participants is not only widely regarded as unethical, but also statistically invalid, since they inherently have much much lower SNR that you cannot see ("unknown unknowns"), and have excessive covariates that are very difficult to isolate (a user "bounce" now can mean not just "uninterested" but also "enraged for something out of your control" or "actively prevented their session from being stitched together", and a "session time" now also includes "researching your vendors and scripts" or "timing errors from client-side mitigations"). 

  This is why success stories of "implicit data" are all from large scale companies with the in house expertise and sheer data volume to compensate for these, they want SMBs and individual site owners to be hyped up with all the "implicit data" gold-rush but the only businesses that can realistically profit from it is, the large companies with, by definition, third-party, data analysts. 
- It's the user's wallet, so the public will eventually catch up to what is happening (first party tracking -> CDP aggregation and stitching -> behavioral data -> more persuasiveness -> more purchases). Even if not everyone have the time and resources to do this, they will still "smell" the difference of heavily "optimized" content and eventually be much more skeptical of the content they are consuming and their true intent.

## The "browser can run scripts and make API requests" argument

Just like any other computer program, having the technical permission at runtime to access a resource on a standardized browser (like installing mutation observers, get the GPU vendor, the user's IP address, etc.), does NOT constitute a legal permission of behavioral or identifying data points to be extracted from such environment and transmitted. 

Some vendors argue "no harm done". I think this is a very weak argument. If I accurately pointed out the battery percentage, operating system, language settings, GPU and OS vendor of a stranger's computer, and then started pitching a product to them (let's say a hairbrush, arguably nothing to do with computers at all), what would they think? Would they be more likely or less likely to purchase the product, or just run for their life?

Additionally trying to detect or circumvent the presence of mitigations (like tracker blockers, or `/proc/self/attr/current` on Linux LSM, or listing the process tree in Windows) without a valid reason is usually enough to be considered malware.

Therefore, I am calling this for what it is from a technical standpoint, a data exfiltration-no euphemisms. It is an **unauthorized data transfer** about the user's computing environment and/or behavior from the user's computer to the vendor's computer, increasingly employing common evasion tools such as CDN cloaking and **encryption/obfuscation** to confuse the custodian of such data (the visitor) of the intent of the transfer. 

## My stance on this

I don't think most "privacy oriented" users (including me) want absolutely anonymity, but rather a promise that data collected from me that is **not used against me and other users to make transactions they would not otherwise have, by virtue of the content or offering alone**. Unfortunately currently the only way to achieve this is to use aggressive blocking and confusion techniques to prevent my data from being collected correctly, attributed correctly and contribute to a more useful "insight". 

While I do acknowledge that there are marketers (and vendors) that use these data very ethically (including the ones mentioned in the ruleset), they themselves are also just part of the loop and the subject of "persuasive" marketing by SSPs, Ad Techs and other upstream vendors. This creates a "supply chain risk" kind of issue where even if your direct vendor is ethical the ultimate behavior of the product might not be. I believe the only rational way to address this risk at the user's end is a "deny first ask for forgiveness later" approach.

I think it really crossed the line that I myself might not be persuaded regardless of whether I get "personalized" for my own visit, but I refuse to contribute data points to this problem that can affect other users and disproportionally vulnerable groups like children, elderly, disabled (i.e., the "optimizations" which is arguably not in my legal power to control once the data is successfully exfiltrated).

## The ruleset

This is an supplementary ruleset for uBlock Origin that I use to block first-party data exfiltration. This will break websites, and is a ruleset that you should use only after all other traditional and arguably more selective venues have been enabled. I will update it regularly as I find more patterns.

```
# First and proxied third-party tracking firewall
#
# ## Preconditions:
#
# This assumes you already have these configured, as they are easier to manage and comes with hard mode:
# * * 3p-frame block
# * * 3p-script block
# no-remote-fonts: * true
# no-csp-reports: * true
#
# DNS tricks and CNAME cloaking is also not in scope, this should be handled by your DNS resolver, they are in a much better position to provide coverage.
# 
# Site storage is also not in scope, they are very difficult to target with high recall and site breakage coming from that can be very difficult to debug. If you are willing to do this, deleting all site data on exit should be acceptable and common practice.
#
# ====== default-deny rules for things that are commonly used exfiltrate data ======

# goal: default deny on common data exfiltration patterns, regardless of intent

*$method=post|put|patch,strict3p,domain=~youtube.com   # normal websites (especially smaller scaled one) don't need to POST data to a different origin
# ref: https://posthog.com/docs/advanced/proxy/nginx
# ref: https://plausible.io/docs/proxy/introduction

*$image,3p,removeparam             # strip data associated with tracking pixel
                                   # we will use 3p to reduce breakage as the amount of data reasonably exfiltrated by a GET request is limited, and retroactive measures are acceptable at this level to me
                                   
*$websocket                        # websockets has to be whitelisted thanks to certain session spying "solutions", and it is usually difficult to audit after the page loaded without sending more data
@@||example.com^$websocket         # whitelist for websocket

# ====== API pattern rules ======
#
# Some platforms inject api requests straight into first party domains
# Goal: block suspiciously named scripts and prevent data exfiltration

||mxpnl.com^$domain=mxpnl.com        # Good on them https://docs.mixpanel.com/docs/session-replay#why-does-it-say-the-player-failed-to-load

||/cdn-cgi/$method=post|put|patch  # Cloudflare
||/cdn-cgi/challenge-platform/*^   # prepend @@ to temporarily allow a challenge

||*/record$method=post|put|patch   # just random ways to say data exfiltration
||*/record$method=post|put|patch
||*/event$method=post|put|patch
||*/events$method=post|put|patch
||*/event/$method=post|put|patch
||*/rum^$method=post|put|patch
||*/collect^$method=post|put|patch

||*/fp.$script         # common variants of saying fingerprintJS
||*/fingerprint$script
||*/collect$script 
||*/rum$script 
||*/rsa$script 
||*/array.js^$script


# Vender specific rules

||*/array/*^                                      # posthog, interesting name choice there

||/api/event^$method=post|put|patch               # plausible, posthog API endpoints that officially endorsed CDN cloaking (automatically lose any argument of "privacy 'respect'")
                                                  # this should already be covered in the general rules but if we temp disable that this is more specific guardrail
||/api/events^$method=post|put|patch              # doesn't hurt
||*/e/^$method=post|put|patch                     # posthog

# ====== Confusion rules  ======
#
# For scripts or tracking pixels that bypassed DNS and API pattern and strict3p filtering
#
# Goal: confuse exfiltrated events to prevent correct attribution

*$removeparam=/^ajs_/ # rudderstack
*$removeparam=idsite  # strip matomo RPC site ID
*$removeparam=_idn
*$removeparam=pv_id  # matomo page view tracker
*$removeparam=cid    # fathom#encodeParameters
*$removeparam=sid    # strip fathom site ID as well
*$removeparam=visitor
```