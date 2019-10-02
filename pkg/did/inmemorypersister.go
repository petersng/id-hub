package did

import (
	didlib "github.com/ockam-network/did"
)

// InMemoryPersister is a persister that stores and get did documents in memory
// Mainly used for testing.
type InMemoryPersister struct {
	store map[string]*Document
}

// GetDocument retrieves a DID document from the given DID
func (p *InMemoryPersister) GetDocument(d *didlib.DID) (*Document, error) {
	if p.store == nil {
		p.store = map[string]*Document{}
	}
	doc, ok := p.store[d.String()]
	if !ok {
		return nil, nil
	}
	return doc, nil
}

// SaveDocument saves a DID document
func (p *InMemoryPersister) SaveDocument(doc *Document) error {
	if p.store == nil {
		p.store = map[string]*Document{}
	}
	p.store[doc.ID.String()] = doc
	return nil
}