---
title: "Misskey Hardening"
description: My own experience and experiments in improving my Misskey instance's security.
date: 2024-10-23T15:35:35-05:00
image: 
math: 
license: 
math: true
hidden: false
comments: true
categories: ["Technical", "Security", "Risk Management"]
tags: ["Technical", "Security", "Misskey", "Risk Management"]
---

## Introduction

This is a summary of my own experience and experiments in improving my Misskey instance's security and some best practices I have been planning on working on. I have split it into two sections: code hardening and system hardening.

This is a preview of my own instance configuration and I will go back to why exactly these features were implemented in the later sections.

### Misskey

My Misskey instance has the following changes made:

- [X] Place a proxy in front of all inboxes for auditing and filtering problematic requests. 
- [X] Disabled all media proxy requests except authenticated drive file uploads.
- [X] Detailed prometheus metrics for slow requests, DB queries, AP processing, failed auths, etc.
- [X] Strict Content Security Policy.
- [X] Strict ActivityPub sanitization by whitelisting properties and normalizing all referential properties.
- [X] Require TLSv1.2+ over port 443 for all ActivityPub requests.

### Media Proxy

My instance have been running a dedicated proxy: [yumechi-no-kuni-proxy-worker](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker), written in Rust and designed to run on both a real server and Cloudflare Workers. 

In order to reach resources on other instances, Misskey uses a proxy system which translates a GET request to another GET request to the target resource. The potential reasons for this design include network stability, CORS and user privacy (this is only my speculation and made no effort to verify this).

However this open proxy design pattern is a well-known vector for abuse. There have been multiple reports of issues related to this component and all are inherent weaknesses that can only be opportunistically mitigated but not completely removed.

My general recommendation is to move the proxy entirely to a dedicated service. 

Unfortunately this is not equivalent to just setting up an external proxy in the configuration. You need to [patch](https://forge.yumechi.jp/yume/yumechi-no-kuni/commit/ec060b7a1461fde92f3069c164af80098d25d6b7) the code to make sure every unauthenticated request is routed through the proxy.


This proxy is designed to be a drop-in replacement for the internal media proxy and is designed to be more configurable, secure and efficient, a couple security-related features include:

- [X] Rate limiting based on both the number of requests and the duration taken to process the request. [sample config](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/branch/main/local.toml)
  (The basic logic is to first "charge" the maximum tier and if the request is completed, "refund" the unused tokens back to the user.)
- [X] A DNS resolver that only resolves to publicly-addressable IPs. [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/commit/aff0fec58f358200a1c24c2daf75e09cfcf9dfe5/src/fetch/mod.rs#L166)
- [X] AppArmor sandboxed server and further sandboxes image processing. This uses the built-in security module of modern Linux kernels to create dynamic sandboxes for different tasks. [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/branch/main/mac/apparmor/yumechi-no-kuni-proxy-worker)
- [X] Fully constrained image parsing with resource limits and hard timeout. [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/branch/main/src/post_process/image_processing.rs)
  (Image processing is moved off the async runtime and with a library that supports resource limits, if parsing still timed out a signal is sent to cancel the thread as a last resort [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/commit/aff0fec58f358200a1c24c2daf75e09cfcf9dfe5/src/lib.rs#L832).) 


## Code Hardening

### Media Proxy - Time-Complexity DoS (Sharkey MR #754)

There have been reports of a time-complexity DoS attack on the media proxy. The initial mitigation provided in the [patch](https://activitypub.software/TransFem-org/Sharkey/-/merge_requests/754/diffs)
caused some issues with some instances (the one I have been interacting with is gib.schmus.is). 

There are a couple issues that I feel this fix may not have fully fixed this issue and thus I opted to move the feature to a dedicated service:

- Generally parsing multi-media files on a process with access to all secrets is not a good idea.
- The internal proxy behavior is actually quite complex and conditionally performs expensive computations such as image processing and conversions. Image parsing is actually a common vector for complexity-based DoS attacks.
  If you inspect the code you will notice the image conversion calls are not guarded with a resource limit or cancellation, potentially allowing some inputs to consume large amount of resources or indefinitely.
- It is normal for many instances to incur hundreds of legitimate requests to the media proxy just by browsing the timeline, while most of them are cheap, if an attacker craft the same number of requests
  with expensive inputs as explained above, it is likely to cause issues.

### Media Proxy - Amplification (GHSA-gq5q-c77c-v236)

Firstly a disclaimer, I am the original author of this report, so expect some conflict of interest that may overestimate the severity of the vulnerability.
I will try to be as objective as possible by always qualifying with required preconditions as much as possible.

The weakness is very simple but sadly not often warned about in tutorials or guides on how to write a proxy. If the input is HTTP and output is also HTTP and the attacker has significant saying on both ends, a malicious user can shape the output of one proxy to be the input of another proxy, we call this a proxy chain. While proxy chains are also common in legitimate deployments, a chain that is crafted by an attacker is likely to cause unexpected behavior. When coupled with the self-discovery and federated nature of `Misskey`, this is particularly dangerous.

Note that a compliant HTTP proxy will never allow requests to loop within each other:

> A proxy MUST send an appropriate Via header field, as described below, in each message that it forwards. An HTTP-to-HTTP gateway MUST send an appropriate Via header field in each inbound request message and MAY send a Via header field in forwarded response messages.
> 
> RFC 9110 HTTP Semantics - 7.6.3 Via

This is not currently done in the Misskey proxy implementation, the `Via` header is entirely stripped from the request completely removing the trace of the request's origin.

The patch I submitted officially to `misskey-dev` only blocked `Misskey`-to-`Misskey` proxy chains due to various reasons in the communications I could not fully bring the official code to compliance.

If you would like to further patch the internal media proxy I suggest these materials to help developing a patch:

[Preventing Malicious Request Loops by Cloudflare)](https://blog.cloudflare.com/preventing-malicious-request-loops/)

[Via header description by MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Via)

[RFC 9110 - HTTP Semantics, 7.6.3 Via](https://httpwg.org/specs/rfc9110.html#field.via)

### ActivityPub Authorization (`GHSA-m2gq-69fp-6hv4`)

Currently the code structure of the ActivityPub section of Misskey did not track the authority of objects throughout, which causes multiple issues involving 
actors presenting objects that are not theirs and the server could be misled into trusting the object.

Unfortunately this is not part of the integrity provided by JSON or HTTP signatures as it is the semantics inside the object that allow the server to trust the object.

The offical patch has been merged on `2024.11.0`, but given the nature of the issue and the patch seemed to be a little whack-a-mole, for my own server I included more stringent checks:
  - Inbox and fetched objects must be "downcasted" to the correct sub-type which will whitelist all the fields that are allowed to be present in the object, making sure every field is validated.
  - Downgrading attacks by sending an HTTP URL within the object is not allowed, it likely would have been allowed with the upstream patch because it only validated the host matches not the scheme.

### Content Security Policy and URL Summary

It is always a good idea to have a strict Content Security Policy when we are rendering remote user-generated content. Currently there is a [cold PR](https://github.com/misskey-dev/misskey/pull/9863) for it.

For example, the URL summarization library ([summaly](https://github.com/syuilo/summaly)), which is used to show the preview of a URL in Notes, has very similar operational semantics to the media proxy but processes HTML instead.

A trivial observation is it takes a [`playerURL`](https://github.com/syuilo/summaly/blob/master/src/general.ts#L37)
  and later was directly [plugged into](https://github.com/misskey-dev/misskey/blob/develop/packages/frontend/src/components/MkUrlPreview.vue#L179) an `iframe` without checking, which will allow arbitrary embedding of content. With CSP so URL must match `frame-src` for it to load.

Unfortunately Misskey was not designed with CSP in mind so a lot of the code is not CSP-friendly, I have implemented a strict one for my own instance but it involved some code changes and undesirable choice between removal of features
such as API docs and broad whitelisting CDNs, which makes it less suitable for a general patch.

My CSP implementation uses the logic of pre-rendering the static scripts injected into the page during initialization and [generating a static SRI](https://forge.yumechi.jp/yume/yumechi-no-kuni/src/branch/master/packages/backend/src/server/csp.ts) to be served with the page.

![CSP Evaluator](/img/2024-12-20_misskey_hardening_csp.png)

If you enter my instance, you will find that you can't even add make a fetch or image tag without going through the proxy. It also mitigates potential HTTP downgrade attacks possible on some non HSTS instances with the `upgrade-insecure-requests` directive.

I also added [configuration parameters](https://forge.yumechi.jp/yume/yumechi-no-kuni/src/commit/6d752fbf0cdbccd3e16b6fda0673b93b818d4f4d/.config/example.yml#L270) where you can overlay your own whitelist (such as for video embeds and CDNs if you want things that intentionally make third-party requests work).


### Email Validation 

The email validation library Misskey uses by default is [`deep-email-validator`](https://github.com/mfbx9da4/deep-email-validator),
despite the name "validator" it was designed more for suggestive correction and not true validation.

For example, the [`validateRegex`](https://github.com/mfbx9da4/deep-email-validator/blob/master/src/regex/regex.ts) actually does not use regex at all:

```typescript {hl_lines=[7]}
export const isEmail = (email: string): string | undefined => {
  email = (email || '').trim()
  if (email.length === 0) {
    return 'Email not provided'
  }
  const split = email.split('@')
  if (split.length < 2) {
    return 'Email does not contain "@".'
  } else {
    const [domain] = split.slice(-1)
    if (domain.indexOf('.') === -1) {
      return 'Must contain a "." after the "@".'
    }
  }
}
```

And it has an [inconsistency](https://github.com/mfbx9da4/deep-email-validator/blob/1745017ecd3c5e41acbc2aafd9a571e2ccb4c615/src/index.ts#L24) in later stage validations.

```typescript {linenostart=14,hl_lines=[9]}
if (options.validateRegex) {
  const regexResponse = isEmail(email)
  if (regexResponse) return createOutput('regex', regexResponse)
}
if (options.validateTypo) {
  const typoResponse = await checkTypo(email, options.additionalTopLevelDomains)
  if (typoResponse) return createOutput('typo', typoResponse)
}
const domain = email.split('@')[1]
if (options.validateDisposable) {
  const disposableResponse = await checkDisposable(domain)
  if (disposableResponse) return createOutput('disposable', disposableResponse)
}
if (options.validateMx) {
  const mx = await getBestMx(domain)
  if (!mx) return createOutput('mx', 'MX record not found')
  if (options.validateSMTP) {
    return checkSMTP(options.sender, email, mx.exchange)
  }
}
```

The same inconsistency was [reused](https://github.com/misskey-dev/misskey/blob/develop/packages/backend/src/core/EmailService.ts#L217) in the Misskey repository meaning blocked domains are also checked incorrectly unfortunately.

From the highlighted lines we can observe that an email address like `me@gmail.com@ <me@blocked.com>` will be allowed for registration. This issue is confirmed to work on the latest version of Misskey and Sharkey,
so for the time being I suggest do not rely on blocking email domains for security purposes. __If you use the email addresses for other purposes please do not trust its formatting and note it can contain more payloads such as SMTP commands.__

There are two patches available for this issue:

- A PR from @sakuhanight which side-loaded a fix: [misskey-dev/misskey#15056](https://github.com/misskey-dev/misskey/pull/15056)
- I have contacted the library author privately and got a green light on a fix, but the library hasn't been updated in years and they are busy with other matters and it is unknown when it will be reviewed. [mfbx9da4/deep-email-validator#92](https://github.com/mfbx9da4/deep-email-validator/pull/92)


## System Hardening

This section discuses some system configurations I used that benefits convenience and security.

### DB Access

When developing new features or debugging issues I often need to refer to the actual production database to see the scale and looks of real data, with docker it is very tempting to just expose the database to the host machine and connect to it directly.
However this opens up a lot of issues, as without extensive configuration the database credential in the config is the root database credential.
While under a private network this is less of an issue, if the database is exposed to the host machine, any process running on the host machine can access it.

To mitigate this I have [set up](https://forge.yumechi.jp/yume/yumechi-no-kuni/src/branch/master/compose_example.yml#L79) an mTLS tunnel to connect to the database securely for development and replication access.
This exposes the redis and postgresql ports to the internet over [mTLS](https://en.wikipedia.org/wiki/Mutual_authentication).

I wrote a tool to handle generating the certificates and setting up the tunnel, it is available [here](https://forge.yumechi.jp/yume/replikey). 
It follows the PKI model which means certificates can be signed or revoked by a CA without notifying every party involved.

```sh
> replikey cert create-ca --valid-days 1825 --dn-common-name "MyInstance Replication Root Certificate Authority" -o ca-certs
Creating CA with options: CreateCaCommand { valid_days: 1825, dn_opts: DnOptSet { dn_country: None, dn_state: None, dn_locality: None, dn_organization: None, dn_organizational_unit: None }, dn_common_name: "MyInstance Replication Root Certificate Authority", output: Some("ca-certs") }
Provide password for CA key:
Repeat password:

> replikey cert create-server --valid-days 365 --dn-common-name "MyInstance Production Server" -d '*.replication.myinstance.com' --ip-address 123.123.123.123 -o server-certs

> replikey cert sign-server-csr --valid-days 365 --ca-dir ca-certs --input-csr server-certs/server.csr --output server-certs-signed.pem

Enter password: 
CSR Params:
Serial number: 7b6a82c3d9171f7ba8fbd8973aac0146dac611dd
SAN: DNS=*.replication.myinstance.com
SAN: IP=123.123.123.123
Not before: 2024-11-02 22:43:56.751788095 +00:00:00
Not after: 2025-11-02 22:43:56.751783366 +00:00:00
Distinguished name: DistinguishedName { entries: {CommonName: Utf8String("MyInstance Production Server")}, order: [CommonName] }
Key usages: [DigitalSignature, DataEncipherment]
Extended key usages: [ServerAuth]
CRL distribution points: []
Do you want to sign this CSR? (YES/NO)
IMPORTANT: Keep this certificate or its serial number for revocation
```

Now I can just run a command and the database is instantly available locally! I can also spawn replication instances on other machines and replicate the database over the internet.

```sh
> replikey network forward-proxy --listen localhost:5432 \
    --sni postgres.replication.myinstance.com --target replication.myinstance.com:8443 \
    --cert client-signed.pem --key test-client/client.key \
    --ca test-ca/ca.pem
```

And of course, while working with production databases, always remember to use a read-only user and limit the access to the minimum required.

### AppArmor

I have applied AppArmor on the media proxy to guarantee the well-defined system effect of the proxy even under memory corruption,
(this class of vulnerability may seem rare but has [already happened](https://github.com/lovell/sharp/security/advisories/GHSA-54xq-cgqr-rpm3) with the library used by Misskey).

AppArmor follows the Mandatory Access Control model (default is deny) and allows dynamic privilege reduction using "subprofiles" and "hats":

Here is a shortened version of my profile to demonstrate its capabilities:

```
profile yumechi-no-kuni-proxy-worker @{prog_path} {
    # This is the initial profile the program was put in
    include <abstractions/base>
    include <abstractions/ssl_certs> # allow access to CA certs
    include <abstractions/apparmor_api/is_enabled>
    include <abstractions/apparmor_api/introspect>
    include <abstractions/apparmor_api/change_profile> # allow using the profile change API
    include <abstractions/openssl>

    # deny all actions that require Linux capabilities
    deny capability,

    # allow loading system libraries
    /{,usr/}lib/**.so.* mr, 
    
    # allow executing itself using the same confinement
    /{,usr/}{,local/}{,s}bin/@{prog} ixr, 

    # allow executing from these paths only if the UID matches the owner of the executable
    owner /var/lib/@{prog}/{,bin}/@{prog} ixr, 

    # Allow reading configuration files
    owner /var/lib/@{prog}/config.toml r,
    /etc/@{prog}/config.toml r,

    network tcp, # allow binding to TCP
    network udp,
    network netlink raw,
    deny network (bind) udp,

    # Allow dropping privileges to the subprofile "serve", this change is irreversible
    change_profile -> yumechi-no-kuni-proxy-worker//serve, 

    profile serve {
        # This is the subprofile the program transitions to 
        # when it has finished initializing and before accepting connections

        include ..., # omitted for brevity

        deny capability,

        # DNS related
        @{etc_ro}/default/nss r,
        @{etc_ro}/protocols r,
        @{etc_ro}/resolv.conf r,
        @{etc_ro}/services r,
        @{etc_ro}/host.conf r,
        ... # omitted for brevity

        network tcp,
        network udp,
        network netlink raw,
        deny network (bind) tcp,
        deny network (bind) udp,

        /{,usr/}{,local/}{,s}bin/@{prog} ixr,
        owner /var/lib/@{prog}/{,bin/}@{prog} ixr,

        # allow sending signals to the image processing hat
        signal (send, receive) set=(int, term, kill) peer=yume-proxy-workers//serve,
        signal (send) set=(int, term, kill, usr1) peer=yume-proxy-workers//serve//image,


        # define a "hat" for image processing
        ^image {
            include <abstractions/base>
            # Allow changing back to the parent profile
            include <abstractions/apparmor_api/change_profile>

            # Deny files (implicit), capabilities, and network.
            deny capability, 
            deny network,

            # Allow receiving signals from the parent profile
            signal (receive) peer=yume-proxy-worker//serve,
        }
    }
}
```

The "hat" mechanism which allows changing permissions without creating new processes is actually very powerful, the design was around this workflow:

1. The program have a thread of work that need to be sandboxed.
2. It generates a 64-bit random number as the token and give this token to the kernel, the kernel records it and applies the sandbox.
3. Later when the program want to exit the sandbox, it must recall the token and have one chance to request the kernel to exit the sandbox. If the token mismatches the process is killed.

My implementation was usually generate a random number, siphash it (which is very fast but very hard to find enough primitives to replicate when you are working with binary attacks) and use it as the token,
and rehash the number to get the token back at the end of the computation.

## Conclusion

Overall these are some measure I have taken to improve my security stance on my Misskey instance. I hope it can be helpful to you as well. 
