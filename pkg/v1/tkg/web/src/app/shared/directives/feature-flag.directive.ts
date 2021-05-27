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

    constructor(
            private templateRef: TemplateRef<any>,
            private viewContainer: ViewContainerRef,
            private appDataService: AppDataService
    ) {
        super();
    }

    ngOnInit() {
        if (this.isEnabled()) {
            this.viewContainer.createEmbeddedView(this.templateRef);
        } else {
            this.viewContainer.clear();
        }
    }

    /**
     * @method isEnabled
     * helper method to retrieve feature flags from AppDataService and set features
     * in directive to be enabled or disabled
     * @returns {any}
     */
    isEnabled() {
        const features = this.appDataService.getFeatureFlags();
        if (features.value == null) {
            return false
        }
        if (features.value['*']) {
            return true;
        }
        return features.value[this.featureToggle];
    }
}
