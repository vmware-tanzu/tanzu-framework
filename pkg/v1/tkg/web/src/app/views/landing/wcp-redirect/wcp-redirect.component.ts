// Angular imports
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

// Third Party Imports
import { takeUntil } from 'rxjs/operators';

// App Imports
import { APP_ROUTES, Routes } from '../../../shared/constants/routes.constants';
import { StepFormDirective } from '../wizard/shared/step-form/step-form';
import { TkgEventType } from '../../../shared/service/Messenger';
import Broker from 'src/app/shared/service/broker';
import { FieldMapUtilities } from '../wizard/shared/field-mapping/FieldMapUtilities';
import { StepMapping } from '../wizard/shared/field-mapping/FieldMapping';

@Component({
    selector: 'tkg-kickstart-ui-wcp-redirect',
    templateUrl: './wcp-redirect.component.html',
    styleUrls: ['./wcp-redirect.component.scss']
})
export class WcpRedirectComponent extends StepFormDirective implements OnInit {

    APP_ROUTES: Routes = APP_ROUTES;
    vcHost: string;
    hasTkgPlus: boolean = false;

    constructor(private router: Router, protected fieldMapUtilities: FieldMapUtilities) {
        super(fieldMapUtilities);
    }

    protected supplyStepMapping(): StepMapping {
        return undefined;
    }

    ngOnInit() {
        Broker.messenger.getSubject(TkgEventType.VC_AUTHENTICATED)
            .pipe(takeUntil(this.unsubscribe))
            .subscribe((data) => {
                this.vcHost = data.payload;
            });
    }

    /**
     * @method navigate
     * @desc helper method to trigger router navigation to specified route
     * @param route - the route to navigate to
     */
    navigate(route: string): void {
        this.router.navigate([route]);
    }

    /**
     * @method launchVsphereWcp
     * @desc helper method to launch vSphere wcp enablement workflow in new window
     */
    launchVsphereWcp() {
        window.open(`https://${this.vcHost}/ui/app/workload-platform/`, '_blank');
    }

    /**
     * @method launchVsphereWcp
     * @desc helper method to launch vSphere wcp enablement workflow in new window
     */
    relaunchVsphereWizard() {

        window.open(`https://${this.vcHost}/ui/app/workload-platform/`, '_blank');
    }
}
