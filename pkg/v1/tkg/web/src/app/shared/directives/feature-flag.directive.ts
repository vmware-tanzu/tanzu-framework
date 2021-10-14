/**
 * Angular imports
 */
import { Directive, Input, TemplateRef, ViewContainerRef, OnInit } from '@angular/core';
import { BasicSubscriber } from '../abstracts/basic-subscriber';
import { AppDataService } from '../service/app-data.service';

/**
 * App imports
 */
@Directive({
    selector: '[featureToggle]'
})
/**
 * @class FeatureToggleDirective
 * directive used for enabling or disabling features by applying directive to
 * HTML markup tags.
 * @usage
 *  FeatureFlagService:
 *   {
 *       features: {
 *          testFeatureA: true,
 *          testFeatureB: false
 *      }
 *  }
 *
 * Feature Toggle Attribute:
 *  <div *featureToggle="'testFeatureA'"> -> Feature is enabled
 *  <div *featureToggle="'testFeatureB'"> -> Feature is disabled
 */
export class FeatureToggleDirective extends BasicSubscriber implements OnInit {
    @Input() featureToggle: string;
    separator = '.';                // used for separating the starting plugin name from the feature
    private negation: boolean;

    constructor(
            private templateRef: TemplateRef<any>,
            private viewContainer: ViewContainerRef,
            private appDataService: AppDataService
    ) {
        super();
    }

    ngOnInit() {
        // caller may use negation by placing '!' as first char; this inverts the display logic
        if (this.usesNegation()) {
            this.negation = true;
            this.featureToggle = this.featureToggle.substr(1);
        }
        if (this.shouldDisplay()) {
            this.viewContainer.createEmbeddedView(this.templateRef);
        } else {
            this.viewContainer.clear();
        }
    }

    private usesNegation() {
        return this.featureToggle.length > 0 && this.featureToggle.charAt(0) === '!';
    }

    private shouldDisplay() {
        if (this.featureToggle.length === 0) {
            console.log('WARNING: Empty feature toggle encountered (recommend remove enclosed elements)');
            return false;
        }
        if (this.featureToggle === '*') {
            console.log('WARNING: Always-on feature toggle encountered (recommend remove toggle directive)');
            return true;
        }
        return this.negation ? !this.isFeatureEnabled() : this.isFeatureEnabled();
    }

    /**
     * @method isFeatureEnabled
     * helper method to retrieve feature flag
     * Returns true if either the global CLI or the specific plugin set 'enabled' for the feature
     * @returns {any}
     */
    isFeatureEnabled() {
        if (this.featureToggle == null) {
            return false;
        }
        let pluginName, featureName;
        const paramArray = this.featureToggle.split(this.separator);
        if (paramArray.length === 1) {
            featureName = paramArray[0];
            return this.appDataService.isCliFeatureFlagEnabled(featureName)
        }
        pluginName = paramArray[0];
        featureName = paramArray[1];
        return this.appDataService.isPluginFeatureFlagEnabled(pluginName, featureName);
    }
}
