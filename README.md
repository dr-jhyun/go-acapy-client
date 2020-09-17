# go-acapy-client

A library for interacting with ACA-py in Go.

## Context

You can create your own Self-Sovereign Identity solution using the Hyperledger Ursa, Indy, Aries stack. Learn more about the background by watching these videos: 

- [The Story of Open SSI Standards](https://www.youtube.com/watch?v=RllH91rcFdE)
- [Decentralized Identifiers (DIDs) - The Fundamental Building Block of Self Sovereign Identity](https://www.youtube.com/watch?v=Jcfy9wd5bZI&)

To become a Aries developer, follow these courses by the Linux Foundation on [edx.org](https://edx.org):

- [Introduction to Hyperledger Sovereign Identity Blockchain Solutions: Indy, Aries & Ursa](https://courses.edx.org/courses/course-v1:LinuxFoundationX+LFS172x+3T2019/course/)
- [Becoming a Hyperledger Aries Developer](https://courses.edx.org/courses/course-v1:LinuxFoundationX+LFS173x+1T2020/course/) 

## Installation

```shell
$ go get -u github.com/ldej/go-acapy-client
```

## Development

Start a local Indy ledger network VON-network. Make a checkout of [github.com/bcgov/von-network](https://github.com/bcgov/von-network). Then run:

```shell script
./manage start --logs
```

This starts 4 Indy nodes and a von-webserver. The von-webserver has a web interface at [localhost:9000](http://localhost:9000) which allows you to browse the transactions in the blockchain.

Start an Aries-Cloud-Agent-Python (ACA-py) instance and configure the right command line parameters. Read about ACA-py and the command line parameters on my blog:

- [Becoming an Aries Developer - Part 1: Terminology](https://ldej.nl/post/becoming-aries-developer-part-1-terminology/)

- [Becoming an Aries Developer - Part 2: Development Environment](https://ldej.nl/post/becoming-aries-developer-part-2-development-environment/)
- [Becoming an Aries Developer - Part 3: Connecting using Swagger](https://ldej.nl/post/becoming-aries-developer-part-3-connecting-using-swagger/)
- [Becoming an Aries Developer - Part 4: Connecting using go-acapy-client](https://ldej.nl/post/becoming-aries-developer-part-4-connecting-using-go-acapy-client/)

## Examples

Examples can be found in the [examples](./examples) folder.

## Implemented Endpoints

### Connections

| Function Name                  | Method                               | Endpoint                                     | Implemented |
|----------------------------------------------|------------------------------|------------------------------|------------------------------|
| QueryConnections            | GET                              | /connections                                 | :heavy_check_mark: |
| GetConnection | GET | /connections/{id} | :heavy_check_mark: |
| CreateInvitation | POST           | /connections/create-invitation               | :heavy_check_mark: |
| ReceiveInvitation | POST          | /connections/receive-invitation              | :heavy_check_mark: |
| AcceptInvitation | POST      | /connections/{id}/accept-invitation          | :heavy_check_mark: |
| AcceptRequest | POST         | /connections/{id}/accept-request             | :heavy_check_mark: |
| RemoveConnection    | POST \<why though :man_facepalming:> | /connections/{id}/remove                     | :heavy_check_mark: |
| SendBasicMessage    | POST                | /connections/send-message                    | :heavy_check_mark: |
| SendPing               | POST               | /connections/send-ping                       | :heavy_check_mark: |

### Schemas

| Function Name  | Method | Endpoint         | Implemented        |
| -------------- | ------ | ---------------- | ------------------ |
| RegisterSchema | POST   | /schemas         | :heavy_check_mark: |
| QuerySchemas   | GET    | /schemas/created | :heavy_check_mark: |
| GetSchema      | GET    | /schemas/{id}    | :heavy_check_mark: |

### Wallet

| Function Name  | Method | Endpoint                         | Implemented        |
| -------------- | ------ | -------------------------------- | ------------------ |
| QueryDIDs      | GET    | /wallet/did                      | :heavy_check_mark: |
| CreateLocalDID | POST   | /wallet/did/create               | :heavy_check_mark: |
| GetPublicDID   | GET    | /wallet/did/public               | :heavy_check_mark: |
| SetPublicDID   | POST   | /wallet/did/public               | :heavy_check_mark: |
| SetDIDEndpoint | POST   | /wallet/set-public-did           | :heavy_check_mark: |
| GetDIDEndpoint | GET    | /wallet/get-public-did           | :heavy_check_mark: |
| RotateKeypair  | PATCH  | /wallet/did/local/rotate-keypair | :heavy_check_mark: |

### Credentials

TODO

### Credential Definitions

TODO

### Issue Credentials

TODO

### Present Proof

TODO

### Revocation

TODO

### Out-of-Band

TODO

### Ledger

TODO

### Server

TODO

### Action Menu

TODO

## Webhooks

When an event occurs in ACA-py, for example a connection request has been received, a webhook is called on your controller on a certain topic. `go-acapy-client` provides a webhook handler where you can register your own functions to handle these events. Based on an event happening you can update your UI or inform the user about the event.

```go
func ConnectionsEventHandler(event acapy.ConnectionsEvent) {
	fmt.Printf("\n -> Connection %q (%s), update to state %q\n", event.Alias, event.ConnectionID, event.State)
}

func BasicMessagesEventHandler(event acapy.BasicMessagesEvent) {
	fmt.Printf("\n -> Received message on connection %s: %s\n", event.ConnectionID, event.Content)
}

func ProblemReportEventHandler(event acapy.ProblemReportEvent) {
	fmt.Printf("\n -> Received problem report: %+v\n", event)
}

r := mux.NewRouter()
webhooksHandler := acapy.WebhookHandler(
    ConnectionsEventHandler,
    BasicMessagesEventHandler,
    ProblemReportEventHandler,
)

r.HandleFunc("/webhooks/topic/{topic}/", webhooksHandler).Methods(http.MethodPost)
```

You are free to choose the URL for your webhooks. Don't forget to set the command-line parameter for ACA-py: `--webhook-url http://localhost:{port}/webhooks`. The URL you provide to ACA-py is the base URL which will be extended with `/topic/{topic}` by default. So whatever URL you choose, make sure that:

- if the `--webhook-url` is `http://myhost:{port}/webhooks` 
- then the webhooks handler should listen on `http://myhost:{port}/webhooks/topic/{topic}`

The `acapy.WebhookHandler` is web framework agnostic and reads the topic from the URL by itself. The handler returned by `acapy.WebhookHandler` has the standard handler signature `func (w http.ResponseWriter, r *http.Request) {}`.