---
title: "Creating an Auditable Account Access System for Temporary Emergencies"
date: 2023-05-07T14:01:18-05:00
categories: ["Technical", "Security", "Risk Management"]
tags: ["Technical", "Security", "Vault", "Personal", "Risk Management"]
---

I found myself in a situation where I have temporary emergencies that I would like my trusted friends to be have audited access to a subset of my accounts with a prescribed short wait period so I could still stay on top of my affairs.

# Objectives

I want to build a system that has the following properties:

- Stores my account credentials securely with API for syncing with Keepass2.
- Identity or Policy based access control.
- A login system that imposes a wait period on these emergency accesses where I could be notified and take action if necessary.
- A WebUI for accessing credentials.

# System design

I saw that the system is complex enough I need a baseline to start with. I chose Hashicorp Vault which is an overkill for the purpose of storing personal account credentials but its built-in WebUI, identity-based access control and plugin capability made this a good match.

We will configure vault with:

- 10 Unseal keys with a key threshold of 3. This will allow my friends to share their key portions generate 
  a root token for me in a permanent situation. [Docs](https://developer.hashicorp.com/vault/tutorials/operations/rekeying-and-rotating)
- Audit log to a file turned on. [Docs](https://developer.hashicorp.com/vault/docs/audit/file)
- Set up a custom authentication plugin that sends email notifications and only allow login after a wait period. [Docs](https://developer.hashicorp.com/vault/docs/plugins/plugin-development)

# System setup

## Code Development

### Password sync

We need to write a little program to sync entries from Keepass2 to Vault.

I tried this ["keepass-vault-sync-plugin"](https://github.com/Orange-OpenSource/keepass-vault-sync-plugin), it didn't work.
I think writing a new Keepass plugin is too much effort for the simplicity of this task so I wrote a standalone application.

The logic is pretty simple, you open the database 
and [translate](https://github.com/eternal-flame-AD/keepass-vault-sync/blob/main/model/entry.go) each entry into a Vault secret.

The source code for this is available [here](https://github.com/eternal-flame-AD/keepass-vault-sync).

**Update 5/10/2023:**

I ended up feeling writing my own Keepass Plugin is much easier from a secret management perspective that
I do not need to find a way to route my Vault AND Keepass secrets to one more place.

I am not very fluent in C# and its build system, only point and click on the GUI of Rider and 
I could not get the [PLGX](https://keepass.info/help/v2_dev/plg_index.html#plgx) plugin system to work.
(Also it does not support .NET 5.0 upwards so I did not want to deal with it).

In the end I designed a system with inspirations from the above mentioned `keepass-vault-sync-plugin` but
fully rewritten with more focus on the features I want,
like being able to do multiple syncs with custom filters based on tag or path.

Here is a screenshot of the plugin in action, 
it will figure out the vault credentials and corresponding filters from special Keepass entries,
sync all entries to vault and optionally delete orphaned entries in vault.

![keepass-vault-sync](https://github.com/eternal-flame-AD/KeepassVaultSync/blob/main/ScreenShots/SyncInProgress.png?raw=true)

Not sure if it is my connection or the Azure storage backend
sometimes I get timeouts on requests so I had to add a retry. So far I have not had a permanent failure after 3 retries.

### Writing the Authentication Plugin

I got a bunch of pretty cheap YubiKeys in the Cloudflare promotion so I was thinking of using YubiKey password login as the authentication method. An additional benefit of this is I can use the key as 2FA method as well.

Vault plugins are just regular binaries the implement a specific IPC interface. It is just much easier to start with an existing plugin instead of trying to write the plugin from scratch. I referenced [this](https://github.com/hashicorp/vault-plugin-auth-azure) in writing this plugin.

In general we need to implement the following paths:

- `auth/emerg-yubiotp/key/(?P<name>.+)` CRUD for the key definitions.
- `auth/emerg-yubiotp/key` list all keys.
- `auth/emerg-yubiotp/login` login with a key.
- `auth/emerg-yubiotp/config` configuration for the plugin. Such as SMTP server, [YubiCloud API key](https://upgrade.yubico.com/getapikey/), etc.

The login part of the code is pretty straightforward. See [here](https://github.com/eternal-flame-AD/vault-auth-emerg-yubiotp/blob/master/path_auth.go)
for the source.

We also need to do some checking on token renew to make sure the key is still eligible for access:

```go
func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	if req.Auth == nil {
		return nil, errors.New("request auth was nil")
	}

	keyName, ok := req.Auth.InternalData["emerg_yubiotp_keyname"].(string)
	if !ok {
		return nil, errors.New("request auth internal key data was nil, try re-authenticating")
	}

	var ks keyState
	entry, err := req.Storage.Get(ctx, "key/"+keyName)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, errors.New("key not found")
	}
	if err := entry.DecodeJSON(&ks); err != nil {
		return nil, err
	}

	if ks.NextEligibleTime < 0 {
		return logical.ErrorResponse("sorry, this key is disabled"), logical.ErrPermissionDenied
	}

	if ks.NextEligibleTime > 0 && time.Now().Unix() > ks.NextEligibleTime {
		return framework.LeaseExtend(30*time.Second, 60*time.Minute, b.System())(ctx, req, d)
	}

	return logical.ErrorResponse("sorry, you are not eligible to renew your lease"), logical.ErrPermissionDenied
}
```

My plugin code is accessible [here](https://github.com/eternal-flame-AD/vault-auth-emerg-yubiotp).

### UI Patching

The built-in vault UI unfortunately does not autodetect additional auth methods and generate forms for them. So I had to patch the UI code to add this authentication method in. The UI does not use TypeScript but ember.js which I am unfamiliar with, but thankfully my use case is simple I can just reference existing code.

After applying the [patch](https://github.com/eternal-flame-AD/vault-auth-emerg-yubiotp/blob/master/ui-patch/vault-ui-auth-emerg-yubiotp.patch) use `make static-dist` to build the UI and `make bin` to build the vault binary.

## Vault Initialization

### Instance Creation

After the plugin and UI patch has been taken care of we will spin up the vault instance.

We will prepare a docker compose file for running vault. We will use the official vault image but we will bring our own vault binary so that the served UI will have the additional authentication method.

```hcl
# config/vault.hcl

/*
 * Vault configuration. See: https://vaultproject.io/docs/config/
 */

storage "azure" {
        accountName = "mfstorvault01"
        accountKey = "xxxxx"
        container = "prod"
}
/*
listener "tcp" {
        address = ":8443"
        tls_disable = 0
        tls_cert_file = "/ssl/fullchain.pem"
        tls_key_file = "/ssl/privkey.pem"
}
*/

listener "tcp" {
        address = ":8080"
        tls_disable = true
}
#log_level = "debug"

disable_mlock = true # https://github.com/mongodb/vault-plugin-secrets-mongodbatlas/issues/22
api_addr = "https://vault.yumechi.jp"
ui = true
plugin_directory = "/vault/plugin"
```

```yaml
# docker-compose.yml
version: "3"

services:
  vault:
    hostname: vault.yumechi.jp
    image: vault
    ports:
      - 4301:8080
      - 4302:8443
    cap_add:
      - IPC_LOCK
    environment:
      - "TZ=America/Chicago"
    volumes:
      - "./config/:/vault/config:rw"
      - "./plugin/:/vault/plugin:rw"
      - "./audit/:/vault/audit:rw"
      - "./bin/vault:/bin/vault:rw" # to mount our own vault binary with modified UI
    command:
      - server
```

Firstly we need to initialize the vault and retrieve the unseal keys and root token.

```sh
vault operator init \
    -key-shares=10 \
    -key-threshold=3
```

Then we could log in on the Web UI and finish setting up my user account there. This is pretty straightforward just enable the `userpass` auth method and create a user. After that I assign a full access policy to my user entity. Then we revoke the root token using the Web UI and log back in using my user account.

### Secret Separation

I created two V1 (non-versioned) KV secret backend at `password/` and `emerg-password/`. The first one will sync fully with my Keepass database and the second one will only sync entries with a tag set.

We will use the [gadget](#password-sync) I wrote earlier to sync the secrets.

```sh
keepass-vault-sync -input=passwords.kdbx -mount=password
keepass-vault-sync -input=passwords.kdbx -mount=emerg-password -tag=emergency-password
```

### Plugin Installation

Copy the built plugin binary to the plugin directory. Then we need to add the binary sha256 into the plugin catalog so the vault server knows the binary is legit.

```sh
vault write sys/plugins/catalog/auth/vault-auth-emerg-yubiotp \
    sha_256=7ee7f4238340cab11152047733ab4e32769664806e10f3440d9f39b45e3461ce \
    command=vault-auth-emerg-yubiotp
```

Then we enable the auth method and write in our first emergency key.

```sh
vault write auth/emerg-yubiotp/config \
    smtp_host=smtp.zoho.eu \
    smtp_from=yume@yumechi.jp \
    smtp_to=yume@yumechi.jp \
    smtp_username=yume@yumechi.jp \
    smtp_password=xxxxxxxx \
    smtp_port=465 \
    yubiauth_client_id=12345 \
    yubiauth_client_key=xxxxxx

vault write auth/emerg-yubiotp/key/somebody \
    alias=somebody \
    public_id=vvxxxxxxx \
    entity_id=ae4e9756-08cd-012e-4aac-03dc2ef191f8 \
    delay=2880 delay_mail=720
```

Go to the Vault UI and it works!

![Vault UI Showing Emergency YubiOTP Authentication Waiting Period Notification](/img/20230507-vault-emerg-login-prompt.jpg)

After verifying that we have received the confirmation email we can speed this up for testing purposes:

```sh
$ vault read auth/emerg-yubiotp/key/somebody
Key                   Value
---                   -----
alias                 somebody-key-1
delay                 10080
delay_mail            720
entity_id             xxxxx
name                  somebody
next_eligible_time    1683523679
public_id             vvxxxxx
$ vault write auth/emerg-yubiotp/key/somebody next_eligible_time=-1 # disable the key
$ vault write auth/emerg-yubiotp/key/somebody next_eligible_time=0 # clear the waiting period
$ vault write auth/emerg-yubiotp/key/somebody next_eligible_time=1 # allow the key to login immediately
Success! Data written to: auth/emerg-yubiotp/key/somebody
```

### Setup Permissions for Emergency YubiOTP

By default only the "default" policy will be assigned to the login which does not permit the key to access any secrets.

I added a policy using the WebUI called `emerg-yubiotp` which will grant the limited secret scope.

```hcl
path "emerg-password/*" {
  capabilities = ["read","list"]
}

path "totp/*" {
  capabilities = ["read", "list"]
}
```

Now when I try to login using the emergency key I can see the scoped secret engines.

![Vault UI Showing Emergency YubiOTP Authentication with Limited Secret Scope](/img/20230507-vault-emerg-ui.png)


# Using the System

## Emergency Sheet

I prepared a document for friends that contains:

1. The objectives of using this system.
1. A short introduction on how to use a Yubikey.
1. 1-2 unseal keys.
1. Step-by-step instructions on how to log in to the Vault UI using the emergency key, including:
    - Instructions on when the Vault is sealed (ask another friend to put their unseal key in as well).
    - How to access credentials.
    - How to access TOTP codes. Unfortunately the Vault UI does not support TOTP yet so we will need to use the Web CLI.

    > Press the terminal icon on top right. Put “list totp/keys” to get a list of accounts. Choose the account you want and then put “read totp/code/something” to get a code for something.
1. Instructions for disaster, how to ask a technically inclined friend to help me in a dire situation:
    - Rebuild my vault instance if my server is lost.
    - Execute the [root token procedure](https://developer.hashicorp.com/vault/tutorials/operations/generate-root) gain full access my vault.


