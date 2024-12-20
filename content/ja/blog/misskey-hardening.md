---
title: "Misskeyハードニング"
description: Misskeyのセキュリティを向上させるための私自身の実験
date: 2024-10-23T15:35:35-05:00
image: 
math: 
license: 
math: true
hidden: false
comments: true
categories: ["Technical", "Security"]
tags: ["Technical", "Security", "Misskey"]
---

（日本語があまり上手ではないので、英語版を読むことをお勧めします。）

## 導入

セキュリティを向上させるための私自身の経験と実験をまとめたいのでこの記事を書きました。２つの方面でセクションに分けていた: コードの強化とシステムの強化です。

これは[私のインスタンス](mi.yumechi.jp)構成のプレビューであり、これらの機能が実装された理由については後のセクションで詳しく説明します。

### Misskey

私のインスタンスのMisskeyは次の変更が加えられています:

- [X] ログおよびフィルタリングするために、inboxの前にプロキシを配置します。
- [X] 認証されたドライブのアップロードを以外のメディアプロキシを無効にしました。
- [X] 遅いリクエスト、DB クエリ、AP 処理、失敗した認証などの詳細な Prometheus メトリック。[@mihari](https://mi.yumechi.jp/@mihari)でいくつかの統計を公開します。
- [X] [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy).
- [X] プロパティをホワイトリストに登録し、すべての参照プロパティを正規化することで、ActivityPub Objectサニタイズします。
- [X] ActivityPub リクエストにポート 443 経由の TLSv1.2+ が必要です。現実的に、オーバーライドを必要とする例はまだ見たことがありません。

### メディアプロキシ

他のインスタンスのリソースにアクセスするために、MisskeyはGETリクエストは別の GETに変換するプロキシを使用します。この設計の潜在的な理由としては、ネットワークの安定性、CORS、ユーザーのプライバシーなどが挙げられます (私の推測にすぎず、検証はしていません)。

ただし、このオープン プロキシ設計パターンは、悪用されるよく知られたベクトルです。このコンポーネントに関連する問題については複数の報告があり、すべて固有の弱点であり、機会を捉えて軽減することしかできず、完全に取り除くことはできません。

一般的に、わたしはプロキシを完全に専用サービスに移動すること勧めます。

残念ながら、これは構成で外部プロキシを設定することと同じではありません。認証されていないすべてのリクエストがプロキシ経由でルーティングされるようにするには、コードに[パッチ](https://forge.yumechi.jp/yume/yumechi-no-kuni/commit/ec060b7a1461fde92f3069c164af80098d25d6b7)を適用する必要があります。


私のインスタンスは専用プロキシ [yumechi-no-kuni-proxy-worker](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker) を実行しています。これは Rust で実装されており、サーバーと Cloudflare Workers の両方で実行されるように設計されています。


このプロキシは、内部メディアプロキシの代替品として設計されており、より構成可能で、安全かつ効率的です。セキュリティ関連の機能には次のものがあります:

- [X] リクエストの数とリクエストの処理にかかった時間の両方に基づくレートリミット。[サンプル構成ファィル](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/branch/main/local.toml)
(基本的なロジックは、最初に最大層を「請求」し、リクエストが完了したら、未使用のトークンをユーザーに「返金」することです。)
- [X] パブリックIPのみDNSリゾルバー。 [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/commit/aff0fec58f358200a1c24c2daf75e09cfcf9dfe5/src/fetch/mod.rs#L166)
- [X] Linux カーネルに組み込まれたセキュリティモジュール (AppArmor) を使用して、さまざまなタスク用の動的なサンドボックスを作成します。[code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/branch/main/mac/apparmor/yumechi-no-kuni-proxy-worker)
- [X] リソース制限とハードタイムアウトによる完全に制約されたイメージ解析。 [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/branch/main/src/post_process/image_processing.rs)
(画像処理はイベントループから移動され、リソース制限をサポートするライブラリを使用してもタイムアウトしている場合は、最後の手段としてスレッドをキャンセルするためのシグナルが送信されます [code](https://forge.yumechi.jp/yume/yumechi-no-kuni-proxy-worker/src/commit/aff0fec58f358200a1c24c2daf75e09cfcf9dfe5/src/lib.rs#L832)。)

## コードの強化

### メディアプロキシ - 時間複雑度 DoS (Sharkey MR #754)

メディアプロキシに対する時間複雑度DoS攻撃の報告がみられる。[パッチ](https://activitypub.software/TransFem-org/Sharkey/-/merge_requests/754/diffs) で提供された最初の緩和策により、一部のインスタンスで問題が発生しました (例は gib.schmus.is )。

この修正ではこの問題が完全には解決されていないと思われる問題がいくつかあるため、この機能を専用サービスに移行することにしました。

- 一般的に、DB Rootとしてアクセスできるプロセスでマルチメディアファイルを解析するのは進められません。
- 内部プロキシの動作は実際には複雑で、画像処理や変換などの高価な計算を条件付きで実行します。画像の解析は、実際には複雑性に基づくDoS攻撃の一般的なベクトルです。コードを調べると、画像変換呼び出しがリソース制限やキャンセルで保護されていないことに気付いた。そのため、一部の入力が大量のリソースを消費したり、無期限に消費したりする可能性があります。
- 多くの場合、タイムラインを参照するだけでメディアプロキシへの正当なリクエストが数百件発生するのは正常です。そのほとんどは安価ですが、攻撃者が上記のように高価な入力を使用して同じ数のリクエストを作成すると、問題が発生する可能性があります。

### Amplification (GHSA-gq5q-c77c-v236)

免責事項として、私はこのレポートのオリジナル作成者であるため、脆弱性の重大性を過大評価する恐れがあります。可能な限り客観的であるよう努めますが、必要な前提条件をできるだけ明確にしておきます。

この脆弱性は非常に単純ですが、残念ながらプロキシの作成方法に関するチュートリアルやガイドで警告されることはあまりありません。入力も出力もHTTPであり、攻撃者が両端で発言権を持っている場合、悪意のあるユーザーは1つのプロキシの出力を別のプロキシの入力にすることができます。これを「プロキシ チェーン」と呼びます。チェーンは正当な展開でも一般的ですが、攻撃者が作成したチェーンは予期しない動作を引き起こす可能性があります。`Misskey` のフェデレーションの性質と組み合わせると、特に危険です。

一般的に、準拠した HTTP プロキシでは、リクエストが互いにループすることは決して許可されない：

> A proxy MUST send an appropriate Via header field, as described below, in each message that it forwards. An HTTP-to-HTTP gateway MUST send an appropriate Via header field in each inbound request message and MAY send a Via header field in forwarded response messages.
>
> （プロキシは、転送する各メッセージで、以下に説明する適切な Viaヘッダーフィールドを送信しなければなりません。HTTP-to-HTTP ゲートウェイは、受信リクエストメッセージごとに適切な Via ヘッダーフィールドを送信しなければならず、転送された応答メッセージに Via ヘッダーフィールドを送信してもよい。）
> 
> RFC 9110 HTTP Semantics - 7.6.3 Via

これは現在、Misskey プロキシ実装では行われていません。`Via` ヘッダーはリクエストから完全に削除され、リクエストの発信元の痕跡が完全に削除されます。

私が正式に `misskey-dev` に提出したパッチは、さまざまな理由で公式コードを完全に準拠させることができなかったため、`Misskey` から `Misskey` へのプロキシ チェーンのみをブロックしました。

内部メディア プロキシをさらにパッチしたい場合は、パッチの開発に役立ってくれると思います。

[Preventing Malicious Request Loops by Cloudflare)](https://blog.cloudflare.com/preventing-malicious-request-loops/)

[Via header description by MDN](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Via)

[RFC 9110 - HTTP Semantics, 7.6.3 Via](https://httpwg.org/specs/rfc9110.html#field.via)

### ActivityPub 認証 (`GHSA-m2gq-69fp-6hv4`)

現在、Misskey の ActivityPub セクションのコード構造は、オブジェクトの権限を全体的に追跡していないため、アクターが自分のものではないオブジェクトを入れて、サーバーがオブジェクトを信頼するように誤解される可能性があるという複数の問題が発生しています。

残念ながら、これはオブジェクト内のセマンティクスによってサーバーがオブジェクトを信頼できるようになるため、JSON または HTTP 署名によって提供される整合性の一部ではありません。

公式パッチは `2024.11.0` にマージされましたが、問題の性質とパッチが少々モグラ叩きのようだったため、自分のサーバーではより厳格なチェックを組み込みました:
  - 受信トレイとフェッチされたオブジェクトは、正しいサブタイプに「ダウンキャスト」する必要があります。これにより、オブジェクト内に存在することが許可されているすべてのフィールドがホワイトリストに登録され、すべてのフィールドが検証されます。
  - オブジェクト内でHTTP URLを送信してダウングレード攻撃を行うことは許可されていません。アップストリームパッチでは、スキームではなくホストの一致のみが検証されていたため、おそらく許可されていたでしょう。


### Content Security PolicyとURLのサマリーサービス

リモートのユーザー生成コンテンツをレンダリングする場合は、厳格なCSPを設定することをお勧めします。現在、これに関する[非アクティブ PR](https://github.com/misskey-dev/misskey/pull/9863) があります。

たとえば、Notes で URL のプレビューを表示するために使用される URL 要約ライブラリ ([summaly](https://github.com/syuilo/summaly)) は、メディアプロキシとよく似た操作セマンティクスを持ちますが、代わりにHTMLを処理します。

簡単に観察すると、このライブラリは [`playerURL`](https://github.com/syuilo/summaly/blob/master/src/general.ts#L37) を受け取り、その後、チェックなしで直接 [iframe` にプラグイン](https://github.com/misskey-dev/misskey/blob/develop/packages/frontend/src/components/MkUrlPreview.vue#L179) され、コンテンツの任意の埋め込みが可能になります。[CSP](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Security-Policy)では、URLが`frame-src`と一致しないとロードされません。

MisskeyはCSPを考慮して設計されていないかもしれないため、多くのコードは CSP フレンドリーではありません。私は自分のインスタンス用に厳密な実装をしましたが、コードの変更がいくつかあり、API ドキュメントや CDN の広範なホワイトリストなどの機能の削除というどっちも望ましくない選択が伴い、一般的なパッチには適さず、テンプレートに適しています。

私の CSP 実装では、初期化中にページに挿入された静的スクリプトを事前レンダリングし、[静的SRIをサーブする](https://forge.yumechi.jp/yume/yumechi-no-kuni/src/branch/master/packages/backend/src/server/csp.ts) してページと共に提供するというロジックを使用しています。
![CSP Evaluator](/img/2024-12-20_misskey_hardening_csp.png)

また、独自のホワイトリストをオーバーレイできる [構成パラメータ](https://forge.yumechi.jp/yume/yumechi-no-kuni/src/commit/6d752fbf0cdbccd3e16b6fda0673b93b818d4f4d/.config/example.yml#L270) も追加しました (ビデオ埋め込みや CDN など、意図的にサードパーティのリクエストを機能させたい場合)。

### メール検証

Misskey がデフォルトで使用するメール検証ライブラリは [`deep-email-validator`](https://github.com/mfbx9da4/deep-email-validator) です。

"validator"という名前にもかかわらず、これは真の検証ではなく、「修正」を目的として設計されています。

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

そして、後の段階の検証では[矛盾](https://github.com/mfbx9da4/deep-email-validator/blob/1745017ecd3c5e41acbc2aafd9a571e2ccb4c615/src/index.ts#L24)があります。

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

Misskey リポジトリでも同じ不整合が [再利用](https://github.com/misskey-dev/misskey/blob/develop/packages/backend/src/core/EmailService.ts#L217) されており、残念ながらブロックされたドメインも誤ってチェックされています。

強調表示された行から、`me@gmail.com@ <me@blocked.com>` のようなメールアドレスが登録に許可されることがわかります。この問題は、Misskey と Sharkey の最新バージョンで動作することが確認されているため、当面は、セキュリティ上の目的でメール ドメインのブロックに依存しないことをお勧めします。 __メールアドレスを他の目的で使用する場合は、フォーマットを信頼しないでください。また、SMTP コマンドなどのペイロードがさらに含まれる可能性があることに注意してください。__

この問題には 2 つのパッチがあります:

- sakuhanight@github さんからのPRで、修正がサイドロードされています: [misskey-dev/misskey#15056](https://github.com/misskey-dev/misskey/pull/15056)
- ライブラリの作者に個人的に連絡し、修正のゴーサインをもらいましたが、ライブラリは何年​​も更新されておらず、作者は他の問題で忙しく、いつレビューされるかは不明です。[mfbx9da4/deep-email-validator#92](https://github.com/mfbx9da4/deep-email-validator/pull/92)


## システムの強化

このセクションでは、利便性とセキュリティにメリットをもたらす、私が使用したいくつかのシステム構成について説明します。

### DB リモートアクセス

新しい機能を開発したり、問題をデバッグしたりするときに、実際のデータの規模や外観を確認するために実際の運用データベースを参照する便利や必要があることがよくあります。Dockerでは、データベースをホストマシンに公開して直接接続したくなるものです。
ただし、詳細な構成を行わないと、`.config` 内のデータベース資格情報がルートデータベース資格情報になるため、多くのリスクが伴います。
プライベートネットワークでは、これはそれほど問題にはなりませんが、データベースがホストに公開されている場合、ホストマシンで実行されているすべてのプロセスがデータベースにアクセスできます。

これを軽減するために、開発とレプリケーション アクセスのためにデータベースに安全に接続するための mTLS トンネルを [設定](https://forge.yumechi.jp/yume/yumechi-no-kuni/src/branch/master/compose_example.yml#L79) しました。

これにより、redis および postgresql ポートが [mTLS](https://en.wikipedia.org/wiki/Mutual_authentication) 経由でインターネットに公開されます。

証明書の生成とトンネルの設定を処理するツールを作成しました。このツールは [こちら](https://forge.yumechi.jp/yume/replikey) から入手できます。

`replikey` は PKI モデルに準拠しており、関係するすべての関係者に通知することなく、CA によって証明書に署名または取り消しを行うことができます。

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

> replikey network forward-proxy --listen localhost:5432 \
    --sni postgres.replication.myinstance.com --target replication.myinstance.com:8443 \
    --cert client-signed.pem --key test-client/client.key \
    --ca test-ca/ca.pem
```

もちろん、データベースで作業するときは、常に読み取り専用ユーザーを使用し、アクセスを必要最小限に制限することをお勧めします。

### AppArmor

メモリ破損時でもプロキシの明確なシステム影響を保証するために、メディアプロキシに AppArmor を適用しました (このクラスの脆弱性はまれにしか発生しないように見えるかもしれませんが、Misskeyが使用するライブラリでは [すでに発生しています](https://github.com/lovell/sharp/security/advisories/GHSA-54xq-cgqr-rpm3))。

「AppArmor」は、必須アクセス制御モデル(デフォルトは拒否)に従い、「サブプロファイル」と「ハット」を使用して動的な権限削減を許可します:

その機能を示すために、私のプロファイルの短縮版を以下に示します:

```
profile yumechi-no-kuni-proxy-worker @{prog_path} {
    # これはプログラムが最初に置かれたプロファイルです
    include <abstractions/base>
    include <abstractions/ssl_certs> # allow access to CA certs
    include <abstractions/apparmor_api/is_enabled>
    include <abstractions/apparmor_api/introspect>
    include <abstractions/apparmor_api/change_profile> # allow using the profile change API
    include <abstractions/openssl>

    # Linuxの機能を必要とするすべてのアクションを拒否
    deny capability,

    # システムライブラリの読み込みを許可する
    /{,usr/}lib/**.so.* mr, 
    
    # 同じ制限を使用してプログラム自体を実行できるようにする
    /{,usr/}{,local/}{,s}bin/@{prog} ixr, 

    # UIDがファイルの所有者と一致する場合にのみ、これらのパスからの実行を許可する
    owner /var/lib/@{prog}/{,bin}/@{prog} ixr, 

    # 設定ファイルの読み取りを許可する
    owner /var/lib/@{prog}/config.toml r,
    /etc/@{prog}/config.toml r,

    network tcp, # TCPへのバインドを許可する
    network udp,
    network netlink raw,
    deny network (bind) udp,

    # サブプロファイル "serve" へ移行のを許可します。元に戻せません。
    change_profile -> yumechi-no-kuni-proxy-worker//serve, 

    profile serve {
        # 初期化が完了し、接続を受け入れる前にこのプロファイルに移行する

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


        # 画像処理用の「ハット」を定義する
        ^image {
            include <abstractions/base>
            # 親プロフィールへの変更を許可する
            include <abstractions/apparmor_api/change_profile>

            # ファイル（暗黙的）、機能、およびネットワークを拒否します。
            deny capability, 
            deny network,

            # 親プロファイルからのシグナルの受信を許可する
            signal (receive) peer=yume-proxy-worker//serve,
        }
    }
}
```

スレッド内で権限を変更できる「ハット」メカニズムは、実際には非常に強力で、設計は次のワークフローに基づいていました。残念ながら、スレッドの依頼ためにこの機能は実行スレッドを制御できる言語 (Go (`runtime.LockOsThread`)、Rust、C など) で使えます。

1. プログラムには、サンドボックス化する必要がある作業のスレッドがあります。

2. 64 ビットの乱数をトークンとして生成し、このトークンをカーネルに渡します。カーネルはそれを記録し、サンドボックスを適用します。

3. 後でプログラムがサンドボックスを終了するときは、トークンを呼び出し、カーネルにサンドボックスを終了するよう要求するチャンスを 1 回与える必要があります。トークンが一致しない場合、プロセスは強制終了されます。

私は通常の実装では、乱数を生成し、それを [siphash](https://en.wikipedia.org/wiki/SipHash) (非常に高速ですが、バイナリ攻撃で作業しているときに複製するのに十分なプリミティブを見つけるのは非常に困難です) してトークンとして使用し、計算の最後にその数値を再ハッシュしてトークンを戻します。

## 結論

これらは、Misskey インスタンスのセキュリティスタンスを改善するために私が行った対策の一部です。皆さんにとっても役立つことを願っています。