package acapy

import "fmt"

func (c *Client) DIDExchangeAcceptInvitation(connectionID string, myEndpoint string, myLabel string) (Connection, error) {
	var connection Connection
	var queryParams = map[string]string{
		"my_endpoint": myEndpoint,
		"my_label":    myLabel,
	}
	err := c.post(fmt.Sprintf("/didexchange/%s/accept-invitation", connectionID), queryParams, nil, &connection)
	if err != nil {
		return Connection{}, err
	}
	return connection, nil
}

func (c *Client) DIDExchangeAcceptRequest(connectionID string, myEndpoint string) (Connection, error) {
	var connection Connection
	var queryParams = map[string]string{
		"my_endpoint": myEndpoint,
	}
	err := c.post(fmt.Sprintf("/didexchange/%s/accept-request", connectionID), queryParams, nil, &connection)
	if err != nil {
		return Connection{}, err
	}
	return connection, nil
}

// >>>>> dr.jhyun ------------------------------------------------------------------------------------------------------
func (c *Client) DIDExchangeCreateRequest(theirPublicDID string, mediationID string, myEndpoint string, myLabel string) (Connection, error) {
	var connection Connection
	var queryParams = map[string]string{
		"their_public_did": theirPublicDID,
		"mediation_id":     mediationID,
		"my_endpoint":      myEndpoint,
		"my_label":         myLabel,
	}
	err := c.post("/didexchange/create-request", queryParams, nil, &connection)
	if err != nil {
		return Connection{}, err
	}
	return connection, nil
}

// <<<<< dr.jhyun ------------------------------------------------------------------------------------------------------
