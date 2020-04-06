package goauth2

import (
	"github.com/inklabs/rangedb"
)

type clientApplication struct {
	IsOnBoarded   bool
	ClientID      string
	ClientSecret  string
	PendingEvents []rangedb.Event
	RedirectUri   string
}

func newClientApplication(records <-chan *rangedb.Record) *clientApplication {
	aggregate := &clientApplication{}

	for record := range records {
		if event, ok := record.Data.(rangedb.Event); ok {
			aggregate.apply(event)
		}
	}

	return aggregate
}

func (a *clientApplication) apply(event rangedb.Event) {
	switch e := event.(type) {
	case *ClientApplicationWasOnBoarded:
		a.IsOnBoarded = true
		a.ClientID = e.ClientID
		a.ClientSecret = e.ClientSecret
		a.RedirectUri = e.RedirectUri

	}
}

func (a *clientApplication) Handle(command Command) {
	switch c := command.(type) {

	case RequestAccessTokenViaClientCredentialsGrant:
		if !a.IsOnBoarded {
			a.Emit(RequestAccessTokenViaClientCredentialsGrantWasRejectedDueToInvalidClientApplicationID{
				ClientID: c.ClientID,
			})
			return
		}

		if a.ClientSecret != c.ClientSecret {
			a.Emit(RequestAccessTokenViaClientCredentialsGrantWasRejectedDueToInvalidClientApplicationSecret{
				ClientID: c.ClientID,
			})
			return
		}

		a.Emit(AccessTokenWasIssuedToClientApplicationViaClientCredentialsGrant{
			ClientID: c.ClientID,
		})

	}
}

func (a *clientApplication) Emit(events ...rangedb.Event) {
	for _, event := range events {
		a.apply(event)
	}

	a.PendingEvents = append(a.PendingEvents, events...)
}

func (a *clientApplication) GetPendingEvents() []rangedb.Event {
	return a.PendingEvents
}