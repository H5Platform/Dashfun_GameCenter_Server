package events

import "github.com/go-telegram/bot/models"

var TGPreCheckoutQueryEvents = NewEvent[*models.PreCheckoutQuery]()
var TGSuccessfulPaymentEvents = NewEvent[*models.Message]()
