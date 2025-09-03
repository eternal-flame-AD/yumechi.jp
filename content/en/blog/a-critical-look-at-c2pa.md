---
title: "A Critical Look at C2PA: DRM and AI"
description: 
date: 2025-05-04T18:04:12-05:00
image: 
math: 
license: 
hidden: false
comments: true
draft: false
categories: ["Technical", "Humanity", "Standards", "C2PA"]
tags: ["Technical", "Standards", "C2PA", "DRM", "AI", "Humanity"]
---

## Foreword

On a broad level, some complex social problems like trust issues, misinformation, privacy, market manipulation, inclusivity, etc. have no good technical solutions. And that is for a good reason, the public does not work like idealized models in technical standards, and one decision can easily make things much worse due to the effect of widespread bias of how the public adopt or uses any new technology. 

Artistic creation transcends technical provenance, and any kind of technical intervention geared to modify public behavior through 'transparency' can inadvertently lead to the intended outcome being unlikely to materialize, and unintended side-effects be amplified regardless of any controls to 'clarify' the intent of the standard. Additionally, any standard is also inevitably influenced by commercial stakeholder interests, and leading to adoptions that are not aligned with the original intent of the standard. 

Third-party cookies is a perfect example of this where the standard Initiative is completely misaligned with industry interests, leading to gross underestimation of the impact of the technology. it is a tell-tale history lesson to learn that technical standards often have wider impact than explicitly intended.

Let's post two questions and keep these in mind as we proceed:

> Will adopting C2PA help or harm my ability to reach my audience?
> 
> Will the public gain meaningful insights into a creative work, or a piece of reporting, that will influence the public appreciation of that work in a positive way?

## Background on C2PA

[C2PA](https://c2pa.org/) is a new standard for encoding cryptographically verifiable metadata, authorship identity assertion and provenance for digital assets.
It has seen some large scale adoptions, such as in [Adobe's Content Authenticity Initiative](https://helpx.adobe.com/creative-cloud/help/cai/adobe-content-authenticity.html).

Firstly, let's paste the official 'Goal and Non-Goals' section from the C2PA spec, to make sure we are on the same page of what it is supposed to do, and my discussions are in no way misrepresenting the standard authors' and stakeholders' objectives (bold marks are mine, but text is unmodified from the spec):

> The goal of the C2PA Specifications for Content Credentials is to tackle the extraordinary challenge of **trusting media** in a context of rapidly evolving technology and the democratization of powerful creation and editing techniques. To this end, the specifications are designed to enable global, **opt-in**, adoption of **digital provenance** techniques through the creation of a rich ecosystem of digital provenance enabled applications for a wide range of individuals, organizations and devices, while meeting appropriate **security** and **privacy** requirements, as well as **human rights** considerations.
> 
> It is important to highlight that Content Credentials do **not provide value judgments about whether a given set of provenance data is 'true'**, but instead merely whether the provenance information can be verified as **associated with the underlying asset, well-formed, and free from tampering**. In addition, the Signer of the Content Credential will be **verified against one or more trust lists**.
>
> [C2PA Explainer - Goal and Non-Goals](https://c2pa.org/specifications/specifications/1.4/explainer/Explainer.html#_goals_and_non_goals)

Additionally, sometimes I quote sentences from adopters of C2PA (such as Adobe and OpenAI), they do not represent the positions of C2PA but are only to note the likely outcome of C2PA adoption, regardless of intent from the standard authors, for example, Adobe's adoption has a slightly different objective than this simple goal of C2PA, and it is geared much more towards protection of copyright, brand, or perceived quality of the product, and that is **a significant deviation from the C2PA standard's goal**.

> Adobe Content Authenticity (Beta) is a web app that allows users to apply, customize, and inspect Content Credentials. It can help you **protect your work** and get **recognition** and allows you to request that generative AI models do not use your work for training or as input to help create new content.  Inspecting Content Credentials with Adobe’s Inspect tool can provide more context about a piece of content, such as who made it and how.
>
> [Adobe Content Authenticity (Beta)](https://helpx.adobe.com/creative-cloud/help/cai/adobe-content-authenticity.html)

> Adobe's Inspect tool allows you to inspect Content Credentials across various media types. Drag and drop a file or screenshot, or use the native file picker to see if the selected content has Content Credentials. The tool allows you to learn more about the selected content, potentially including **who was involved in creating it**, **how it was created**, and **whether generative AI was used**. 
>
> [Adobe's Inspect tool](https://helpx.adobe.com/creative-cloud/help/cai/adobe-content-authenticitiy-inspect.html)

## Provenance of Media Assets

Provenance means an assertion of the chronological (history of) custody of an asset ([Wikipedia](https://en.wikipedia.org/wiki/Provenance)), such as :
- In the software industry, provenance generally means a manifest of all creators, entities and projects that have contributed to the final product (directly through writing code, indirectly through a dependency, or even an academic publication outlining the methodology or algorithm used), and information regarding their licensing terms or restrictions. 

  For example 'This software is copyrighted by ACME Inc., all rights reserved, with code licensed under the MIT license from John Doe' can be part of a provenance manifest.
- In C2PA, provenance generally encodes the usage of a 'Content Credential aware tool', to create an assertion that says a particular asset has been manipulated by a particular tool, and this assertion can be cryptographically verified from the final artwork.

  For example, a hypothetical C2PA provenance manifest might encode the assertion that 'This artwork was created using GIMP, version 2.10, by John Doe, on 2025-01-01', and optionally it can even encode usage of specific tools, like embedded AI models, etc.

Now let's analyze how does this help C2PA achieve its goals, and potential of side-effects that might be undesirable.

This is often quoted as 'combating misinformation' (against both human generated or deepfake AI generated content). This is a very broad topic, and will immediately hit some theoretical barriers regardless of technical implementation.

### Public Education

It has been widely known that theoretical trust models almost never work in practice. Cryptographic proof of identities, both free and non-free, both anonymous or tied to government issued IDs, have existed in a long time. However, success rate of phishing and identity theft never goes down.

In an ideal world, everyone can get a PGP smartcard for anonymous identities, and obtain an S/MIME certificate that asserts their government issued ID. This would create an in theory leak-proof system where nobody can claim to be you or any of your asserted identities. However this is not how it works in practice. Even if you have perfect security hygiene, you are still at risk on both ends:
  - You can certainly be impersonated, if someone call up the pharmacy and claim to be you, they are likely to presume that the caller is indeed you. That is why further verification is always required and placed in hard to bypass guidelines in the organizational policies. This is not possible to apply to internet media sources where the public is not required or is going in to in any meaningful capacity verify the content is 'authentic' by using any 'inspect tool' that shows chronology of creation.
  - You can also certainly impersonate others, as long as you claim to be them, or are in an authority position to override any question of identity, this is the basis of many urgency based scams (e.g. claiming they are from the police and are out to get you unless you perform irreversible actions right away).
  
Let's evaluate how C2PA plays with this framework.

- If the public cannot be educated to verify strings of cryptographic signatures, how is the public going to be able to verify any asset using the 'inspect tool' is not tampered with, especially it doesn't really "judge whether a given set of provenance data is 'true'", and the public still have to do their own due diligence which is very hard if not impossible to do?
- What counts as 'trustworthy' in the context of C2PA? It is well-known logical and information theory limit that (1) **if you want to prove something that not everyone else can prove, you must demonstrate control over some private information that only satisfactory parties own (like the software used to make the assertion)**, (2) **if the verifier cannot independently verify the assertion, validation of any assertion requires implicit trust in a surrogate assertion, such as 'The US government must only issue passports to US citizens, thus if this passport is genuine (which is independently verifiable), then the person holding it must be a US citizen'**, this manifests as:

  You can never prove the authenticity of any assertion that a specific tool was used without some sort of obfuscation (i.e. DRM). If I develop my own drawing software, I should be allowed to assert authorship and provenance of my artwork, and it should carry the same weight as any professional tool. Even without considering any financial motives, an idealistic trusted entity in C2PA would never sign any open source software or hardware, simply because anyone can just modify it and assert fake provenance data. While some provenance solutions never intended it to be cryptographically secure against modification, cryptographic security IS an explicitly mentioned goal of C2PA. This is not just an abstract thought experiment, competitive drawing software like [Krita](https://krita.org/) and [GIMP](https://www.gimp.org/) already exist, and they should not be excluded from the "trusted" list.

  While There have been some theoretical advances that tries to work around this (Zero-knowledge proofs, for example), but they did not fundamentally break this limit. Additionally provenance (attestation that something happened in the past) is a much more difficult problem to tackle than simply proving the authenticity of a current state. This is exactly why attestation in cryptographic products always only work for closed platforms, such as smartcards, HSMs and processor enclaves. To prove that any particular asset is created using a particular software, you must reword the assertion to not break this theoretical limit, and that means hide the key used to sign the attestation from public view and within the private sphere of that tool.

#### Counter-argument

Some people may intuitively say "hey Yumechi I think you are being cornered into a narrow application of C2PA, but C2PA is an open standard and, and how to create your own keys for your own tool is certainly possible". Okay, let's read the spec wording:

> To enable consumers to make informed decisions about the provenance of
> an asset, and prevent unknown attackers from impersonating others, it is
> critical that each application and ecosystem reliably identify the
> owner of the signing credential (also known as a digital certificate) is
> issued. A certification authority (CA) performs this real-world due
> diligence to ensure signing credentials are only issued to verified
> entities. CAs that are recognized and trusted in a specific application
> or ecosystem are included in a trust list, which is a list of
> certification authorities that issue signing credentials for that
> application.
>
> [C2PA Explainer - Trust](https://c2pa.org/specifications/specifications/1.4/explainer/Explainer.html#_trust)

How is that not prescribing a de facto standard for a 'trust list' (of authorized entities to vet artist or tool origins) that is used by all C2PA applications?
While the C2PA standard allows for decentralized or user-generated trust models in theory, in practice, adoption is likely to coalesce around a small set of centralized authorities, much like existing PKI systems.

It doesn't really matter whether one can theoretically create a decentralized trust network, any technical solution's strength and drawbacks are dictated by practical applications. If we only analyze technologies with what is theoretically possible, people can call PKI better than PGP in every aspect because PKI technically can be used where everybody is their own CA and thus it works as well as PGP, doesn't work like that reductionist analysis. 


### What even is Misinformation?

Now let's shift gears into the heart of the matter, what is misinformation? What is their goal? What is their critical vulnerability we can exploit? Does the C2PA theoretical model, in its current form, reliably do that?

While some people intuitively think of misinformation as deepfake users that create fake images, videos, or audio, then fabricate an entire story around it so the public can't tell it apart from reality, the truth is misinformation is more subtle than that, and **effective misinformation campaigns rarely involve fabrication of media or events out of thin air**. The goal of misinformation is never to make the public believe one specific event happened when it did not, but rather to subtly shift the public's view toward an abstract concept, be it a political figure, a company, or any other entity. C2PA only tackles the problem of verifying the media assets are authentic, but has very limited efficacy in combating subtle misinformation through narrative manipulation, or selective reporting of facts.

And this view is reflected in empirical observations, generally "effective" mitigations involve **reputation based systems**, like classifying news sources against their factual correctness and political bias, social media moderation or neutral public education campaigns, rather than specific technical mitigations such that in theory the public can 'vet' any media asset independently, which is a very high bar to expect the public to be able to do. Even if a theoretical 'inspect tool' is available, the amount of knowledge required to use the tool is so high that it is not practical to use it.

Also we need to be wary of the pitfall of **resolution of trust issues versus shifting the trust from one entity to another**, while I am not accusing any particular organization, dependence of any proprietary tool to vet the authenticity of media assets is in itself a source of bias and potentially places one in the position to be influenced by censorship. We shouldn't magically decide 'news sources are not to be trusted and they need more accountability' but now just simply shift that trust to vendors of proprietary tools or trust anchors to vet the authenticity of media assets in an entirely unbiased way. 

In reality, bias certainly will happen, people under economic hardship will be excluded from this "golden seal" of trust, and it would certainly be explicitly or implicitly devalued from the social media or web presence as a whole if this technology becomes widely adopted.

### What is Trust

One might argue, but C2PA did clarify that it is not trying to find "true" media, but just provides provenance data that may be helpful. More information is better than less information, right? Unfortunately, this plays again into the imperfect nature of human trust, and irrelavant proof can be worse than no proof at all. Here, whether a photo is actually taken by a real camera, or modified by a commercial tool intended for human use, is _not_ relevant to any goal of C2PA.

"Too much crypto" usually refers to using cryptography of excessive strength for a purpose, however, there is a more harmful form of "too much crypto" that is not often discussed, especially in the realm of assertions, attestations and proofs, that is providing proof that is irrelevant to the context.

C2PA did clarify it didn't try to find "true" medias out of fake ones. It is technically correct but it seems to imply an assumption that there is a "more trusted" state when there is not. No proof of image authenticity or tool usage will increase trust in an irrelevant context (authentic art or imagery or narrative).

In cryptography a effective proof requires multiple orders of magnitude higher effort than potential gain for an adversary to defeat a system for the proof to be 'secure'. Anything worse than that, that proof generally is as good as nothing, possession of that proof is disregarded in any evaluation of trust. **This system is trivially defeated by an adversary for the purpose of combating misinformation or asserting human artistic authenticity**, regardless of the strength of the crypto system itself, they just need to add misleading provenance data or put the image in a larger misleading context and highlight the C2PA proof.

What makes it harmful is a machine can be programmed to disregard a signal presented to them completely, humans are heuristic machines and cannot regardless of how hard they try, they will:

- over-compensate like from not trusted to distrust: "why are you pushing C2PA? I refuse to invest a cent into the ecosystem and I am cancelling my subscription" or
- under-compensate like from not trusted to more trusted: "this articles doesn't sound right.. but it has C2PA from a reputable news agency so I guess it is more true than that article"

#### Bank Fraud Example

To put it into a concrete example, a fraudster cashing a fake check usually produces excessive proof, like multiple forms of ID, a long story to explain to the clerk how they obtained the check, etc. Even if none of these assertions provide any real proof on the legitimacy of the check, a naive bank clerk is likely to reduce their suspicion. Similarly, a misinformation campaign can also use C2PA against itself, by providing a "legitimate" and "waterproof" attestation on the legitimacy of the asset, to hide the actual bias of the content. 

## Authenticity and Fine Art

It is apparent that many adopters of C2PA does not truly use it to combat misinformation, but to use it 'off-label' to assert some sort of authenticity to assets produced by artists in their ecosystem. This goes much less into the "fact-checking" aspect but well into the realm of **appreciation of fine art**, and creates critical concern over C2PA's efficacy in that regard.

C2PA mentions that all use of AI models is logged in tool usage metadata:

> How does C2PA address the use of AI/ML in the creation and editing of assets?
>
> Each action that is performed on an asset is recorded in the asset’s Content Credentials. These actions can be performed by a human or by an AI/ML system. When an action was performed by an AI/ML system, it is clearly identified as such through it’s digitalSourceType field.
>
> [C2PA Explainer - Use of Artificial Intelligence/Machine Learning (AI/ML)](https://c2pa.org/specifications/specifications/1.4/explainer/Explainer.html#_use_of_artificial_intelligence_machine_learning_aiml)


The appreciation of fine art is a very subjective, intricate, and human-centric topic, and reducing it to the chronological sequence of tool usage is a gross simplification of the creative process. This creates not only abstract philosophical problems but also practical concerns over ethical issues and human rights.

### What is an Identity?

Any artistic creation (writing, drawing, music or even programming) is a product of the human mind, and regardless of the market value of the final product, that creation deserves evaluation of human ingenuity from the creator, not any material presentation of the creation (how expensive the tool is, the 'identity' of the creator, etc.).

Open source developers has over a long history adopted an pseudonym culture over identity of their work, and it has shown very good results in practice and worked around many practical limitations that anonymity provides:

1. **Creativity transcends identity**: Pseudonym culture is a very effective way to freely create and share ideas without concern of professional repercussions or 'ruining' one's reputation. A work posted under a pseudonym deserve the same respect over work posted under a legal name (be it tied to a birth certificate with the government, or an account with a company, it IS the same thing). I am free to publish my work under any name I want, if I sign 'Da Vinci' on my work and become established in my community, then users of my work must address me as 'work from Da Vinci', period. That is human decency to respect people's work by the name they are known by, and prefer to be addressed by. Does that mean people would really think it is Leonardo's work and thus make it more valuable? No. The value of Leonardo's work is not on the signature of the canvas, but the ingenuity of the creation, which is superimposed on the final canvas. Tying an identity to a creative work does nothing in establish "value" of any work.
2. **Most projects do not need legal litigation power**: While submitting over under a pseudonym is certainly problematic if you require legal litigation power, most projects do not need that, and generally speaking, for true commercial projects where legal IP protection that can stand up to a formal lawsuit is needed, the product would not be free, and that litigation power is manifested as a licensing agreement, service agreement, or a non-disclosure agreement, never simply by the name of the company written on the code.
3. **Creative Work have intrinsic self-provenance**: When interviewers ask about an open source project the applicant maintains, they never meant 'flip out the private key or photo ID so I know you cryptographically signed the bytes in the commit history', they always meant tell me about the project, what you did, what you are proud of, why you made specific choices that are unexpected to the interviewer, etc. 
   
   Deep insights into creative process is very difficult, if not impossible to fake, and this is the essence of self-provenance: we do not require any external authority to vouch for the authenticity of the work, we vouch for it by ourselves, by encoding unique human insight, emotion, and experience into the work, and demonstrating that we hold the intrinsic value that enabled us to make this creation rather than the stream of tokens on the source code. 
   
   An encoding of tool usage and signature from a CA is, in my opinion, a shallow and reductionist take on what truly identifies a creative work, after all one can already do that by buying an S/MIME certificate and a timestamp service, likely costing much less than a subscription for Creative Cloud, and they can cancel their account without invalidating previous signatures.
4. **Visibility does not have to coincide with Centralization**: While people rely on GitHub, Adobe Marketplace, etc. to make their work visible, developers do _not_ rely on Microsoft (Github) to vouch for their developer identity. We sign our work independently using decentralized PGP keys, and rely on self-provenance to make a name for ourselves. If I delete my GitHub account, move my work to a different platform or my own website, I can continue to use that established identity. On the contrary, if one can no longer prove they are the same creator when they canceled their Creative Cloud subscription, how is it a robust solution for properly attributing any creative work to the creator?

### What is AI Usage?

While the C2PA standard simply took a neutral stance and encoded boolean flags for AI training consent, and a technical solution to track the use of AI model in any asset. I believe the true impact of such change must be critically evaluated, and whether under public perception of the value of AI in creative work, this can truly deliver a net positive.

Firstly, it is widely known fact that generally the public associates "AI usage" in creative work with "low quality", "low effort", "devoid of human touch", etc. This set a very real stage for the practical meaning of these labels, regardless of the intent of the standard authors or even technology adopters. However this is a simplification, while one may claim the output for a generic prompt, from a publicly available model to be "low effort", it is not the case for all creative work.

Both historically and in the modern world of fine art creation, artists come in all shapes and sizes, some cannot afford proper equipment, some have a busy schedule and do not have time to draw every single line, some used to be first-class artists but developed debilitating hand tremors, some have mental and physical disabilities that make it impossible to materialize their full creative potential. Generative AI models may be their only way of artistic expression, if they put in the effort to layout a specific backstory, composition and color choices, how isn't the final product worthy of appreciation as result of human ingenuity? Forcing them to sign their work off as AI involved is dismissing the abstract concept that motivated their creation, and in my opinion, the only really meaningful way to evaluate the value of any creative work.

Secondly, if we truly reduce the creative process into the iterative, chronological sequence of tool usage, we are not only dismissing the creative impulse, but ironically start to look suspiciously reminiscent of the high level idea of all generative AI models: generate a convincing (valued by human) output from a random source, by applying a sequence of transformations while minimizing the amount of 'unnaturalness' (loss) of the output. It is _Reductio ad absurdum_ to require a declaration of AI usage (and implicit encouragement of segregation of AI generated work), while simultaneously modeling the creative process exactly like a generic generative AI model.

Empirically to draw a parallel, in the open source community, no one is requiring declaration of tools used to create a commit (be it a keyboard, an AI model, an IDE or a text editor, it would be a disaster and inviting for off topic discussions!). The reason is simple, the value of any submission is not in the process that yielded syntax tree that materialized the logic, but the insight, logic and architecture of the commit. While it is certain that people _also_ associate "AI generated code" with "bad code", just by not proving an option to filter out "AI generated code", the platform is providing a fair evaluation and appreciation of any contribution, regardless of the tool used to make it.

## Pitfall of Voluntary and Opt-in

While C2PA is a voluntary standard, we must acknowledge that any 'voluntary' standard has broad implications on stakeholders, adopters, individual users and the public at large, regardless of whether they personally made the choice to adopt it or not, **the action to apply the technology or not is voluntary, but the felt impact of the technology is not.** 

Simply objecting to C2PA is not enough in counteracting the potential devaluation of your work, and systematic shift in the public perception of values of creative work even if they choose not to adopt it.

I encourage creators to take a step back and evaluate holistically the real impact of C2PA or industry variants. In my opinion, the insights offered in C2PA (tool usage, creator identity) are superficial, potentially misleading, unlikely to be accessed or understood by many, and irrelevant to the deeper appreciation of creative intent or journalistic truth. 