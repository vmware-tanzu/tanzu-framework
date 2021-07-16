// Angular imports
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

// App imports
import { takeUntil } from "rxjs/operators";
import { APP_ROUTES } from '../../constants/routes.constants';
import Broker from "../../service/broker";
import { TkgEvent, TkgEventType } from "../../service/Messenger";
import { BasicSubscriber } from "../../abstracts/basic-subscriber";
import { EditionData } from "../../service/branding.service";

/**
 * @class HeaderBarComponent
 * HeaderBarComponent is the Clarity header component for TKG Kickstart UI.
 */
@Component({
    selector: 'tkg-kickstart-ui-header-bar',
    templateUrl: './header-bar.component.html',
    styleUrls: ['./header-bar.component.scss']
})
export class HeaderBarComponent extends BasicSubscriber implements OnInit {

    edition: string = '';
    docsUrl: string = '';

    constructor(private router: Router) {
        super();
    }

    ngOnInit() {
        Broker.messenger.getSubject(TkgEventType.BRANDING_CHANGED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data: TkgEvent) => {
                const content: EditionData = data.payload;
                this.edition = content.edition;
                this.docsUrl = (this.edition === 'tkg') ? 'https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html' :
                    'http://tanzucommunityedition.io/docs';
            });
    }

    /**
     * @method navigateHome
     * helper method to route user to application home route
     */
    navigateHome() {
        this.router.navigate([APP_ROUTES.LANDING]);
    }

    navigateToDocs() {
        window.open(this.docsUrl, "_blank");
    }
}
