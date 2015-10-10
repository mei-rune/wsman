package wsman

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/runner-mei/wsman/envelope"
)

func Subscribe(ep *Endpoint, ns, name string, selectorSet map[string]string,
	deliveryMode string,
	endToAddress string,
	endToIdentifier string,
	recvAddress string,
	recvIdentifier string,
	queryList map[string][]envelope.QueryFilter,
	sendBookmarks bool) (io.Reader, *xml.Decoder, error) {
	input := &envelope.Subscribe{
		Namespace:   ns,
		MessageId:   Uuid(),
		Name:        name,
		SelectorSet: selectorSet,

		DeliveryMode:    deliveryMode,
		EndToAddress:    endToAddress,
		EndToIdentifier: endToIdentifier,
		RecvAddress:     recvAddress,
		RecvIdentifier:  recvIdentifier,
		QueryList:       queryList,
		SendBookmarks:   sendBookmarks}
	reader, err := ep.Deliver(strings.NewReader(input.Xml()))
	if nil != err {
		return nil, nil, err
	}
	//defer closeReader(reader)

	decoder := xml.NewDecoder(reader)
	if err = ReadEnvelopeResponse(decoder, "SubscribeResponse"); nil != err {
		closeReader(reader)
		return nil, nil, err
	}
	return reader, decoder, nil
}

func SubscribeByPull(ep *Endpoint, ns, name string, selectorSet map[string]string,
	queryList map[string][]envelope.QueryFilter,
	sendBookmarks bool) (*Enumerator, error) {
	reader, decoder, e := Subscribe(ep, ns, name, selectorSet, envelope.DELIVERYMODE_WSMAN_PULL, "", "", "", "", queryList, sendBookmarks)
	if nil != e {
		return nil, e
	}

	defer closeReader(reader)
	ok, e := locateElement(decoder, "EnumerationContext")
	if nil != e {
		return nil, e
	}
	if !ok {
		return nil, ElementNotExists("Envelope/Body/SubscribeResponse/EnumerationContext")
	}

	context, e := readXmlText(decoder)
	if nil != e {
		return nil, e
	}

	return &Enumerator{Namespace: ns,
		Endpoint:    ep,
		Name:        name,
		SelectorSet: selectorSet,
		Context:     context,
		is_pull:     true}, nil
}
