package config

import "fmt"

type checkBoxWizardStep struct {
	wiz      *wizard
	modal    *checkBoxModalLayout
	showImpl func(*checkBoxModalLayout)
}

func newCheckBoxStep(wiz *wizard, currentStep int, totalSteps int, helperText string, width int, title string, showImpl func(*checkBoxModalLayout), done func(map[string]bool), back func(), pageID string) *checkBoxWizardStep {

	step := &checkBoxWizardStep{
		wiz:      wiz,
		showImpl: showImpl,
	}

	title = fmt.Sprintf("[%d/%d] %s", currentStep, totalSteps, title)

	// Create the modal
	modal := newCheckBoxModalLayout(
		step.wiz.md.app,
		title,
		width,
		helperText,
	)

	modal.done = done
	modal.back = back
	step.modal = modal

	page := newPage(nil, pageID, "Config Wizard > "+title, "", modal.borderGrid)
	step.wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

	return step
}

func (step *checkBoxWizardStep) show() {
	step.showImpl(step.modal)
}
