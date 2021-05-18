import { Directive } from '@angular/core';
import { ClrTooltipTrigger } from '@clr/angular';
​
/**
 * Adds "role" and "aria-haspopup" attributes to Clarity clr-icon tag
 */
@Directive({
    selector: '[a11yTooltipTrigger]',
    host: {
        tabindex: '0',
        'aria-label': 'tooltip',
        '[class.tooltip-trigger]': 'true',
        '[attr.aria-describedby]': 'ariaDescribedBy',
        '[attr.role]': '"tooltip"',
        '[attr.aria-haspopup]': '"true"',
        '[attr.aria-label]': '"tooltip"'
    }
})
export class A11yTooltipTriggerDirective extends ClrTooltipTrigger { }
