package config

import "fmt"

type textBoxWizardStep struct {
	wiz      *wizard
	modal    *textBoxModalLayout
	showImpl func(*textBoxModalLayout)
}

func newTextBoxWizardStep(wiz *wizard, currentStep int, totalSteps int, helperText string, width int, title string, labels []string, maxLengths []int, regexes []string, showImpl func(*textBoxModalLayout), done func(map[string]string), back func(), pageID string) *textBoxWizardStep {

	step := &textBoxWizardStep{
		wiz:      wiz,
		showImpl: showImpl,
	}

	title = fmt.Sprintf("[%d/%d] %s", currentStep, totalSteps, title)

	// Create the modal
	modal := newTextBoxModalLayout(
		step.wiz.md.app,
		title,
		width,
		helperText,
		labels,
		maxLengths,
		regexes,
	)

	modal.done = done
	modal.back = back
	step.modal = modal

	page := newPage(nil, pageID, "Config Wizard > "+title, "", modal.borderGrid)
	step.wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

	return step
}

func (step *textBoxWizardStep) show() {
	step.showImpl(step.modal)
}
