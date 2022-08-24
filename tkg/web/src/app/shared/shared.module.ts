/**
 * Angular modules
 */
import { NgModule }     from '@angular/core';
import { CommonModule } from "@angular/common";
import { FormsModule, ReactiveFormsModule } from "@angular/forms";
import { HttpClientModule } from '@angular/common/http';

/**
 * Third party imports
 */
import { ClarityModule } from "@clr/angular";

/**
 * App imports
 */
import { FeatureToggleDirective } from './directives/feature-flag.directive';
import { A11yTooltipTriggerDirective } from './directives/a11y-tooltip-trigger.directive';
import { RemoveAriaLabelledByDirective } from './directives/remove-aria-labelledBy.directive';
import { ContextualHelpComponent } from './contextual-help/contextual-help.component';
import { KeyboardListenerDirective } from './directives/keyboard-listener.directive';

const declaredAndExportedModules = [
    CommonModule,
    ClarityModule,
    FormsModule,
    ReactiveFormsModule,
    HttpClientModule
];

/**
 * Module for shared UI components
 */
@NgModule({
    imports: [
        ...declaredAndExportedModules,
    ],

    providers: [],
    exports: [
        ...declaredAndExportedModules,
        FeatureToggleDirective,
        A11yTooltipTriggerDirective,
        RemoveAriaLabelledByDirective,
        ContextualHelpComponent,
        KeyboardListenerDirective
    ],
    declarations: [
        FeatureToggleDirective,
        A11yTooltipTriggerDirective,
        RemoveAriaLabelledByDirective,
        ContextualHelpComponent,
        KeyboardListenerDirective
    ]
})
export class SharedModule { }
