// Angular imports
import { Component } from '@angular/core';
import { Router } from '@angular/router';

// App imports
import { APP_ROUTES } from '../../constants/routes.constants';

/**
 * @class HeaderBarComponent
 * HeaderBarComponent is the Clarity header component for TKG Kickstart UI.
 */
@Component({
    selector: 'tkg-kickstart-ui-header-bar',
    templateUrl: './header-bar.component.html',
    styleUrls: ['./header-bar.component.scss']
})
export class HeaderBarComponent {

    constructor(private router: Router) {

    }

    /**
     * @method navigateHome
     * helper method to route user to application home route
     */
    navigateHome() {
        this.router.navigate([APP_ROUTES.LANDING]);
    }

    goToLink(url: string) {
        window.open(url, "_blank");
    }
}
