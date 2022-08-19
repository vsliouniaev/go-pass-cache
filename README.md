[![Go Report Card](https://goreportcard.com/badge/github.com/vsliouniaev/go-pass-cache)](https://goreportcard.com/report/github.com/vsliouniaev/go-pass-cache)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/vsliouniaev/go-pass-cache?sort=semver)](https://github.com/vsliouniaev/go-pass-cache/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/vsliouniaev/pass-cache?color=blue)](https://hub.docker.com/r/vsliouniaev/pass-cache/tags)

_Better than sending passwords directly through your instant messaging client!_

## Motivation

Security is a compromise between convenience and safety. Sending passwords over
instant-messaging applications (Slack, Skype, Teams, etc) is endemic, particularly in
smaller companies, where security tooling tends to rank lower.

Using multiple separate channels to send data improves security - for example, sending an encrypted email with the decryption key sent through SMS. In recent years this has started to become more common.

## Approach

### Features
* Allow at-most-once access to the data I want to share:
  * Data is erased immediately as it is accessed.
* Limit the time the data is accessible for:
  * Data is inaccessible after 5m.
* Don't trust the server
  * Server has no knowledge of encryption keys- all data is encrypted and decrypted in the end-user's browser using a [libsodium](https://github.com/tonyg/js-nacl)
  * Key is never sent to the server by using a URL fragment.
* Send the data through multiple routes:
  * Only encrypted data is stored on the server at a random URL
  * URL and password are send to the target user through an instant-messaging client
* Fast end-user experience
  * Perform encryption, generate the URL and copy it with `Shift + Return` or one button tap
  * Paste the URL automatically fetches, decrypts and displays the data
  * Everything works from one URL

### How it works
#### Sender
1. The user visits the website (passcache.net), and types in the password they want to share with another user.
2. The brower auto-generates an **Id** and symmetric encryption **Key** directly in the browser on the user's computer. The encryption **Key** never leaves the user's compuater.
3. The user-provided data is encrypted using the encryption **Key** and sent to the website's server under the generated **Id**.
4. A URL is constructed locally in the browser using the **Id** and appended with the **Key** as a _[fragment](https://en.wikipedia.org/wiki/URI_fragment)_. For example https://passcache.net?97e82ccc#912e, the Id is `97e82ccc` and the Key is `912e`.
5. The user sends the full URL through their instant-messaging application.

#### Receiver

1. The receiver pastes the URL into their address bar.
2. The browser sends everythig before the _fragment_ to the server, which looks up the encrypted data using the **Id**, deletes it from memory, and sends it back to the browser. The _fragment_ portion of the URL, which is the **Key** necessary for decryption, is not sent to the back-end server.
3. The browser decrypts the encrypted data using the **Key** from the _fragment_ and shows it to the user.


#### Diagram


```
=== SENDER ===
                                                 ┌─────────────┐
                                                 │passcache.net│
                                                 │             │
                                                 │             │
                                                 │  ▲    ▲     │
                                                 │  │    │     │
                         ┌───────────────────┐   └──┼────┼─────┘
                         │ Browser           │      │    │
                         │                   │      │    │
                         │ ┌───────────────┐ │      │    │
                         │ │Id  (Generated)├─┼──────┘    │
                         │ └───────────────┘ │           │
                         │                   │    ┌──────┴──┐
         ┌────────┐      │ ┌───────────────┐ │    │Encrypted│
         │Password├──────┤►│Key (Generated)├─┼───►│Password │
         └────────┘      │ └───────────────┘ │    └─────────┘
                         │                   │
                         └───────────────────┘



          https://passcache.net/get?Id(generated)#Key(generated)


=== RECEIVER ===

┌───────────────┐
│ passcache.net │
│               │
│    ┌───────┐  │
│    │       │  │          ┌─────────────────────┐
└────┼───────┼──┘          │  Browser            │
     │       │             │                     │
     │       │             │  ┌──────────────┐   │
     │       └─────────────┼─ │Id  (From URL)│   │
     │                     │  └──────────────┘   │
     ▼                     │                     │
 ┌─────────┐               │  ┌──────────────┐   │   ┌────────┐
 │Encrypted├───────────────┼─►│Key (From URL)├───┼──►│Password│
 │Password │               │  └──────────────┘   │   └────────┘
 └─────────┘               │                     │
                           └─────────────────────┘
```


## Notes

* First version written in July 2013 in C#, and pre-dates [Firefox send](https://support.mozilla.org/en-US/kb/send-files-anyone-securely-firefox-send) by a few years.
* Original codebase is at https://github.com/vsliouniaev/pass-cache.
* Thanks to @eoincampbell for suggesting the URL fragment trick.
