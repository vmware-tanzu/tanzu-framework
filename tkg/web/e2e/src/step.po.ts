import { element, by } from 'protractor';

export abstract class Step {

    getNextButton() {
        return element(by.buttonText('next'));
    }

    /**
     * Selects an option based on containing text
     * @param selectElement the 'select' element
     * @param text the 'text' to be matched for
     */
    selectOptionByText(selectElement, text) {
        selectElement.element(by.cssContainingText('option', text)).click();
    }

    /**
     * Selects an option based on its index
     * @param selectElement the 'select' element
     * @param optionIndex the index of the option to be selected, starting with 1.
     */
    selectOptionByIndex(selectElement, optionIndex) {
        selectElement.element(by.css(`option:nth-child(${optionIndex})`)).click();
    }

    /**
     * Selects an option in a ng-select
     * https://github.com/ng-select/ng-select/issues/1270
     * @param ngSelectName the formcontrolname of the element
     * @param optionIndex the index of the item to be selected
     */
    ngSelectOptionByIndex(ngSelectName, optionIndex) {
        element(
            by.css('[formcontrolname="' + ngSelectName + '"] span.ng-arrow-wrapper')
        ).click();

        element(
            by.css(
                '[formcontrolname="' +
                ngSelectName +
                '"] .ng-option:nth-child(' +
                optionIndex +
                ')'
            )
        ).click();
    }

    /**
     * Selects an option in a datalist
     * @param selectName the formcontrolname of the element
     * @param optionIndex the index of the item to be selected
     */
    selectDatalistByIndex(selectName, optionIndex) {
        const input = element(
            by.css(`input[formcontrolname=${selectName}]`)
        );
        input.click();

        const value = element(
            by.css(`datalist#${selectName} option:nth-child(${optionIndex})`)
        ).getAttribute("value");

        input.sendKeys(value);
    }

        /**
     * Selects an option in a datalist
     * @param selectName the formcontrolname of the element
     * @param optionText the text of option value to be inputed
     */
    selectDatalistByText(selectName, optionText) {
        const input = element(
            by.css(`input[formcontrolname=${selectName}]`)
        );
        input.click();

        input.sendKeys(optionText);
    }

    /**
     * Determines if the flow has reached to this step
     */
    abstract hasMovedToStep();
}
