package config

import "fmt"

type choiceWizardStep struct {
	wiz      *wizard
	modal    *choiceModalLayout
	showImpl func(*choiceModalLayout)
}

func newChoiceStep(wiz *wizard, currentStep int, totalSteps int, helperText string, names []string, descriptions []string, width int, title string, direction int, showImpl func(*choiceModalLayout), done func(int, string), back func(), pageID string) *choiceWizardStep {

	step := &choiceWizardStep{
		wiz:      wiz,
		showImpl: showImpl,
	}

	title = fmt.Sprintf("[%d/%d] %s", currentStep, totalSteps, title)

	// Create the modal
	modal := newChoiceModalLayout(
		step.wiz.md.app,
		title,
		width,
		helperText,
		names,
		descriptions,
		direction,
	)

	modal.done = done
	modal.back = back
	step.modal = modal

	page := newPage(nil, pageID, "Config Wizard > "+title, "", modal.borderGrid)
	step.wiz.md.pages.AddPage(page.id, page.content, true, false)
	modal.page = page

	return step
}

func (step *choiceWizardStep) show() {
	step.showImpl(step.modal)
}
