package config

import (
	"github.com/rivo/tview"
)

type page struct {
    parent *page
    id string
    title string
    description string
    content *tview.Box
}

func newPage(parent *page, id string, title string, description string, content *tview.Box) *page {

    return &page{
        parent:  parent,
        id: id,
        title: title,
        description: description,
        content: content,
    }

}

func (p *page) getHeader() string {

    header := p.title
    parent := p.parent
    for {
        if parent == nil {
            return header
        }
        header = parent.title + " > " + header
        parent = p.parent
    }

}