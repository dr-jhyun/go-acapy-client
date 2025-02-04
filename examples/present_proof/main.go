package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type App struct {
	client         *acapy.Client
	server         *http.Server
	ledgerURL      string
	tailsServerURL string
	port           int
	label          string
	seed           string
	rand           string
	myDID          string

	connection             acapy.Connection
	schema                 acapy.Schema
	credentialDefinitionID string
	credentialExchange     acapy.CredentialExchangeRecord
	presentationExchange   acapy.PresentationExchangeRecord
	revocationRegistry     acapy.RevocationRegistry
}

func (app *App) ReadCommands() {
	scanner := bufio.NewScanner(os.Stdin)

	didResponse, err := app.RegisterDID(app.label, app.label+app.rand)
	if err != nil {
		app.Exit(err)
	}
	app.myDID = didResponse.DID
	fmt.Printf("Hi %s, your registered DID is %s\n", app.label, didResponse.DID)

	for {
		fmt.Println(`Choose:
	(1) Create invitation
	(2) Receive invitation
	(3) Register schema
	(4) Create credential definition
	(5) Issue credential
	(6) Send presentation proposal
	(7) Send presentation request
	(8) Send presentation
	(9) Verify presentation
	(10) List presentation proofs
	(exit) Exit
`)
		fmt.Print("Enter Command: ")
		scanner.Scan()
		command := scanner.Text()

		switch command {
		case "exit":
			app.Exit(nil)
			return
		case "1":
			fmt.Println("Who/What is the invitation for?")
			scanner.Scan()
			theirLabel := scanner.Text()

			invitationResponse, err := app.client.CreateOutOfBandInvitation(
				acapy.CreateOutOfBandInvitationRequest{
					Alias:              theirLabel,
					HandshakeProtocols: acapy.DefaultHandshakeProtocols,
					MyLabel:            app.label,
				},
				true,
				false,
			)
			if err != nil {
				app.Exit(err)
			}
			invitation, _ := json.Marshal(invitationResponse.Invitation)
			fmt.Printf("Invitation json: %s\n", string(invitation))
		case "2":
			fmt.Println("Invitation json: ")
			scanner.Scan()
			invitation := scanner.Bytes()
			connection, err := app.ReceiveInvitation(invitation)
			if err != nil {
				app.Exit(err)
			}
			app.connection = connection
			fmt.Printf("Connection ID: %s\n", app.connection.ConnectionID)
		case "3":
			fmt.Print("Schema name: ")
			scanner.Scan()
			schemaName := scanner.Text()

			fmt.Printf("Version: ")
			scanner.Scan()
			version := scanner.Text()

			fmt.Printf("Attributes (comma separated, i.e.: name,age): ")
			scanner.Scan()
			attributesString := scanner.Text()
			attributes := strings.Split(attributesString, ",")

			app.schema, err = app.RegisterSchema(schemaName, version, attributes)
			if err != nil {
				app.Exit(err)
			}
			fmt.Printf("Schema: %+v\n", app.schema)
		case "4":
			fmt.Println("This is slow, it takes a couple of seconds.")
			app.credentialDefinitionID, err = app.client.CreateCredentialDefinition("tag", true, 10, app.schema.ID)
			if err != nil {
				app.Exit(err)
			}
			fmt.Printf("Credential Definition ID: %s\n", app.credentialDefinitionID)
		case "5":
			fmt.Printf("Comment: ")
			scanner.Scan()
			comment := scanner.Text()

			var attributes []acapy.CredentialPreviewAttribute

			for _, attr := range app.schema.AttributeNames {
				fmt.Printf("Attribute %q value: \n", attr)
				scanner.Scan()
				value := scanner.Text()
				attributes = append(attributes, acapy.CredentialPreviewAttribute{
					Name:     attr,
					MimeType: "text/plain",
					Value:    value,
				})
			}

			if credentialExchange, err := app.client.IssueCredential(
				app.connection.ConnectionID,
				acapy.NewCredentialPreview(attributes),
				comment,
				app.credentialDefinitionID,
				app.myDID,
				app.schema.ID,
			); err != nil {
				app.Exit(err)
			} else {
				app.credentialExchange = credentialExchange
			}

		case "6":
			fmt.Printf("Comment: ")
			scanner.Scan()
			comment := scanner.Text()

			var attributes []acapy.PresentationAttribute

			mimeTypes, _ := app.client.CredentialMimeTypes(app.credentialExchange.Credential.Referent)
			for attrName, attrValue := range app.credentialExchange.Credential.Attributes {
				attributes = append(attributes, acapy.PresentationAttribute{
					Name:                   attrName,
					CredentialDefinitionID: app.credentialExchange.CredentialDefinitionID,
					MimeType:               mimeTypes[attrName],
					Value:                  attrValue,
					Referent:               app.credentialExchange.Credential.Referent,
				})
			}

			proposal := acapy.PresentationProposalRequest{
				Comment:             comment,
				AutoPresent:         false,
				PresentationPreview: acapy.NewPresentationPreview(attributes, nil),
				ConnectionID:        app.connection.ConnectionID,
				Trace:               false,
			}
			presentationExchange, err := app.client.SendPresentationProposal(proposal)
			if err != nil {
				app.Exit(err)
			}
			app.presentationExchange = presentationExchange
		case "7":
			fmt.Printf("Comment: ")
			scanner.Scan()
			comment := scanner.Text()

			requestedAttributes := map[string]acapy.RequestedAttribute{}

			for _, attr := range app.presentationExchange.PresentationProposalDict.PresentationProposal.Attributes {
				requestedAttribute, _ := acapy.NewRequestedAttribute(
					nil,
					attr.Name,
					nil,
					acapy.NonRevoked{
						From: time.Now().Add(-time.Hour * 24 * 7).Unix(),
						To:   time.Now().Unix(),
					},
				)
				requestedAttributes[attr.Name] = requestedAttribute
			}

			request := acapy.PresentationRequestRequest{
				Trace:        false,
				Comment:      comment,
				ConnectionID: app.connection.ConnectionID,
				ProofRequest: acapy.NewProofRequest(
					"Proof request",
					"1234567890",
					nil,
					requestedAttributes,
					"1.0",
					&acapy.NonRevoked{
						From: time.Now().Add(-time.Hour * 24 * 7).Unix(), // One week ago
						To:   time.Now().Add(time.Hour * 24 * 7).Unix(),  // One week ahead
					},
				),
			}

			presentationExchange, err := app.client.SendPresentationRequestByID(app.presentationExchange.PresentationExchangeID, request)
			if err != nil {
				app.Exit(err)
			}
			app.presentationExchange = presentationExchange
		case "8":
			// What about the Revealed flag? -> in case of multiple credentials
			requestedAttributes, _, err := app.client.FindMatchingCredentials(app.presentationExchange.PresentationRequest)

			proof := acapy.NewPresentationProof(requestedAttributes, nil, nil)

			presentationExchange, err := app.client.SendPresentationByID(app.presentationExchange.PresentationExchangeID, proof)
			if err != nil {
				app.Exit(err)
			}
			app.presentationExchange = presentationExchange
		case "9":
			presentationExchange, err := app.client.VerifyPresentationByID(app.presentationExchange.PresentationExchangeID)
			if err != nil {
				app.Exit(err)
			}
			app.presentationExchange = presentationExchange
		case "10":
			credentials, err := app.client.GetPresentationCredentialsByID(app.presentationExchange.PresentationExchangeID, 0, "", nil, 0)
			if err != nil {
				app.Exit(err)
			}
			for _, credential := range credentials {
				fmt.Printf("Credential %s: %+v\n", credential.CredentialInfo.Referent, credential.CredentialInfo.Attrs)
			}
		}
	}
}

func (app *App) StartACApy() {
	id := strings.Replace(app.label+app.rand, " ", "_", -1)
	cmd := exec.Command("aca-py",
		"start",
		"--auto-provision",
		"-it", "http", "0.0.0.0", strconv.Itoa(app.port+1),
		"-ot", "http",
		"--admin", "0.0.0.0", strconv.Itoa(app.port+2),
		"--admin-insecure-mode",
		"--genesis-url", fmt.Sprintf("%s/genesis", app.ledgerURL),
		"--seed", app.seed,
		"--endpoint", fmt.Sprintf("http://localhost:%d/", app.port+1),
		"--webhook-url", fmt.Sprintf("http://localhost:%d/webhooks", app.port),
		"--label", app.label,
		"--public-invites",
		"--wallet-type", "indy",
		"--wallet-name", id,
		"--wallet-key", id,
		"--auto-accept-invites",
		"--auto-accept-requests",
		"--auto-ping-connection",
		"--auto-respond-credential-proposal",
		"--auto-respond-credential-offer",
		"--auto-respond-credential-request",
		"--auto-store-credential",
		"--tails-server-base-url", app.tailsServerURL,
		"--preserve-exchange-records",
	)
	cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	go func() {
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (app *App) StartWebserver() {
	r := mux.NewRouter()
	webhookHandler := acapy.CreateWebhooksHandler(acapy.WebhookHandlers{
		ConnectionsEventHandler:          app.ConnectionsEventHandler,
		ProblemReportEventHandler:        app.ProblemReportEventHandler,
		CredentialExchangeEventHandler:   app.CredentialExchangeEventHandler,
		RevocationRegistryEventHandler:   app.RevocationRegistryEventHandler,
		PresentationExchangeEventHandler: app.PresentationExchangeEventHandler,
		CredentialRevocationEventHandler: app.CredentialRevocationEventHandler,
		OutOfBandEventHandler:            app.OutOfBandEventHandler,
	})

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
	})
	r.HandleFunc("/webhooks/topic/{topic}/", webhookHandler).Methods(http.MethodPost)
	fmt.Printf("Listening on http://localhost:%d\n", app.port)
	fmt.Printf("ACA-py Admin API on http://localhost:%d\n", app.port+2)

	app.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", app.port),
		Handler: r,
	}

	go func() {
		if err := app.server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func (app *App) Exit(err error) {
	if err != nil {
		log.Println("ERROR:", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func (app *App) ConnectionsEventHandler(event acapy.Connection) {
	if event.Alias == "" {
		connection, _ := app.client.GetConnection(event.ConnectionID)
		event.Alias = connection.TheirLabel
	}
	app.connection = event
	fmt.Printf("\n -> Connection %q (%s), update to state %q rfc23 state %q\n", event.Alias, event.ConnectionID, event.State, event.RFC23State)
}

func (app *App) CredentialExchangeEventHandler(event acapy.CredentialExchangeRecord) {
	connection, _ := app.client.GetConnection(event.ConnectionID)
	app.credentialExchange = event
	fmt.Printf("\n -> Credential Exchange update: %s - %s - %s\n", event.CredentialExchangeID, connection.TheirLabel, event.State)
}

func (app *App) RevocationRegistryEventHandler(event acapy.RevocationRegistry) {
	app.revocationRegistry = event
	fmt.Printf("\n -> Revocation Registry update: %s - %s\n", event.RevocationRegistryID, event.State)
}

func (app *App) ProblemReportEventHandler(event acapy.ProblemReportEvent) {
	fmt.Printf("\n -> Received problem report: %+v\n", event)
}

func (app *App) PresentationExchangeEventHandler(event acapy.PresentationExchangeRecord) {
	app.presentationExchange = event
	connection, _ := app.client.GetConnection(event.ConnectionID)
	fmt.Printf("\n -> Presentation Exchange update: %s - %s - %s\n", connection.TheirLabel, event.PresentationExchangeID, event.State)
}

func (app *App) CredentialRevocationEventHandler(event acapy.CredentialRevocationRecord) {
	fmt.Printf("\n -> Issuer Credential Revocation: %s - %s - %s\n", event.CredentialExchangeID, event.RecordID, event.State)
}

func (app *App) OutOfBandEventHandler(event acapy.OutOfBandEvent) {
	fmt.Printf("\n -> Out of Band Event: %q state %q\n", event.InvitationID, event.State)
}

func main() {
	var port = 4455
	var name = ""
	var ledgerURL = "http://localhost:9000"
	var tailsServerURL = "http://localhost:6543"

	flag.IntVar(&port, "port", 4455, "port")
	flag.StringVar(&name, "name", "Alice", "alice")
	flag.Parse()

	acapyURL := fmt.Sprintf("http://localhost:%d", port+2)

	app := App{
		client:         acapy.NewClient(acapyURL),
		ledgerURL:      ledgerURL,
		tailsServerURL: tailsServerURL,
		port:           port,
		label:          name,
		rand:           strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)),
	}
	app.StartWebserver()
	app.ReadCommands()
}

func (app *App) RegisterDID(alias string, seed string) (acapy.RegisterDIDResponse, error) {
	didResponse, err := acapy.RegisterDID(
		app.ledgerURL+"/register",
		alias,
		seed,
		acapy.Endorser,
	)
	if err != nil {
		log.Fatal(err)
		return acapy.RegisterDIDResponse{}, err
	}
	app.label = alias
	app.seed = didResponse.Seed
	app.StartACApy()
	return didResponse, nil
}

func (app *App) RegisterSchema(name string, version string, attributes []string) (acapy.Schema, error) {
	schema, err := app.client.RegisterSchema(
		name,
		version,
		attributes,
	)
	if err != nil {
		log.Printf("Failed to register schema: %+v", err)
		return acapy.Schema{}, err
	}
	fmt.Printf("Registered schema: %+v\n", schema)
	return schema, nil
}

func (app *App) ReceiveInvitation(inv []byte) (acapy.Connection, error) {
	var invitation acapy.OutOfBandInvitation
	err := json.Unmarshal(inv, &invitation)
	if err != nil {
		return acapy.Connection{}, err
	}
	return app.client.ReceiveOutOfBandInvitation(invitation, true)
}
